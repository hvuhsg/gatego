package gatego

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/hvuhsg/gatego/config"
	"github.com/hvuhsg/gatego/contextvalues"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const serviceName = "gatego"

type GateGo struct {
	config config.Config
	ctx    context.Context
}

func New(ctx context.Context, config config.Config, version string) *GateGo {
	ctx = contextvalues.AddVersionToContext(ctx, version)
	return &GateGo{config: config, ctx: ctx}
}

func (gg GateGo) Run() error {
	useOtel := gg.config.OTEL != nil
	if useOtel {
		otelConfig := otelConfig{
			ServiceName:             serviceName,
			SampleRatio:             gg.config.OTEL.SampleRatio,
			CollectorTimeout:        time.Second * 5, // TODO: Add to config
			TraceCollectorEndpoint:  gg.config.OTEL.Endpoint,
			MetricCollectorEndpoint: gg.config.OTEL.Endpoint,
			LogsCollectorEndpoint:   gg.config.OTEL.Endpoint,
		}
		shutdown, err := setupOTelSDK(gg.ctx, otelConfig)
		if err != nil {
			return err
		}
		defer shutdown(context.Background())
	}

	// Gather checks and create checker
	checker := createChecker(gg.config.Services)
	checker.Start()

	table, err := NewHandlersTable(gg.ctx, useOtel, gg.config.Services)
	if err != nil {
		return err
	}

	server := gg.createServer(table)
	defer server.Shutdown(gg.ctx)

	serveErrChan, err := serve(server, gg.config.SSL.CertFile, gg.config.SSL.KeyFile)
	if err != nil {
		return err
	}

	// Wait for interruption.
	select {
	case err = <-serveErrChan:
		return err
	case <-gg.ctx.Done():
		return server.Shutdown(context.Background())
	}
}

func createChecker(services []config.Service) *Checker {
	checker := &Checker{Delay: 5 * time.Second, OnFailure: func(err error) {}}

	for _, service := range services {
		for _, path := range service.Paths {
			for _, checkConfig := range path.Checks {
				check := Check{
					Name:    checkConfig.Name,
					Cron:    checkConfig.Cron,
					URL:     checkConfig.URL,
					Method:  checkConfig.Method,
					Timeout: checkConfig.Timeout,
					Headers: checkConfig.Headers,
				}

				checker.Checks = append(checker.Checks, check)
			}
		}
	}

	return checker
}

func (gg GateGo) createServer(table HandlerTable) *http.Server {
	mux := http.NewServeMux()

	// handleFunc is a replacement for mux.HandleFunc
	// which enriches the handler's HTTP instrumentation with the pattern as the http.route.
	handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		// Configure the "http.route" for the HTTP instrumentation.
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		mux.Handle(pattern, handler)
	}

	handleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler := table.GetHandler(r.Host, r.URL.Path)

		if handler == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		handler.ServeHTTP(w, r)
	})

	// Add HTTP instrumentation for the whole server.
	handler := otelhttp.NewHandler(mux, "/")

	addr := fmt.Sprintf("%s:%d", gg.config.Host, gg.config.Port)

	// Start HTTP server.
	server := &http.Server{
		Addr:         addr,
		BaseContext:  func(_ net.Listener) context.Context { return gg.ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      handler,
	}

	return server
}

func serve(server *http.Server, certfile *string, keyfile *string) (chan error, error) {
	supportTLS, err := checkTLSConfig(certfile, keyfile)
	if err != nil {
		return nil, err
	}

	serveErr := make(chan error, 1)

	go func() {
		if supportTLS {
			log.Default().Printf("Serving proxy with TLS %s\n", server.Addr)
			serveErr <- server.ListenAndServeTLS(*certfile, *keyfile)
		} else {
			log.Default().Printf("Serving proxy %s\n", server.Addr)
			serveErr <- server.ListenAndServe()
		}
	}()

	return serveErr, nil
}

func checkTLSConfig(certfile *string, keyfile *string) (bool, error) {
	if keyfile == nil || certfile == nil || *keyfile == "" || *certfile == "" {
		return false, nil
	}

	if !fileExists(*keyfile) {
		return false, fmt.Errorf("can't find keyfile at '%s'", *keyfile)
	}

	if !fileExists(*certfile) {
		return false, fmt.Errorf("can't find certfile at '%s'", *certfile)
	}

	return true, nil
}

func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)

	if os.IsNotExist(err) {
		return false
	}

	// If we cant check the file info we probably can't open the file
	if err != nil {
		return false
	}

	return true
}
