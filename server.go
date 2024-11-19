package gatego

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/hvuhsg/gatego/internal/config"
	"github.com/hvuhsg/gatego/internal/oauth"
	"github.com/hvuhsg/gatego/pkg/multimux"
)

type gategoServer struct {
	*http.Server
}

func newServer(ctx context.Context, config config.Config, useOtel bool) (*gategoServer, error) {
	multimuxer, err := newMultiMuxer(ctx, config.Services, useOtel)
	if err != nil {
		return nil, err
	}

	rootHandler := newRootHandler(multimuxer, config.OAuth)

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	// Start HTTP server.
	server := &http.Server{
		Addr:         addr,
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      rootHandler,
	}

	return &gategoServer{Server: server}, nil
}

func newRootHandler(multimuxer *multimux.MultiMux, oauthConfig *oauth.OAuthConfig) http.Handler {
	rootHanlder := http.NewServeMux()

	// Custom configured services
	rootHanlder.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		multimuxer.ServeHTTP(w, r)
	})

	// Handle OAuth if configured
	if oauthConfig != nil {
		rootHanlder.Handle(oauthConfig.BaseURL, oauth.NewOAuthHandler(*oauthConfig))
	}

	return rootHanlder
}

func newMultiMuxer(ctx context.Context, services []config.Service, useOtel bool) (*multimux.MultiMux, error) {
	mm := multimux.NewMultiMux()

	for _, service := range services {
		for _, path := range service.Paths {
			handler, err := NewHandler(ctx, useOtel, service, path)
			if err != nil {
				return nil, err
			}

			mm.RegisterHandler(service.Domain, path.Path, handler)
		}
	}

	return mm, nil
}

func (gs *gategoServer) serve(certfile *string, keyfile *string) (chan error, error) {
	supportTLS, err := checkTLSConfig(certfile, keyfile)
	if err != nil {
		return nil, err
	}

	serveErr := make(chan error, 1)

	go func() {
		if supportTLS {
			log.Default().Printf("Serving proxy with TLS %s\n", gs.Addr)
			serveErr <- gs.ListenAndServeTLS(*certfile, *keyfile)
		} else {
			log.Default().Printf("Serving proxy %s\n", gs.Addr)
			serveErr <- gs.ListenAndServe()
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
