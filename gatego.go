package gatego

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hvuhsg/gatego/config"
)

type GateGo struct {
	config config.Config
}

func New(config config.Config) *GateGo {
	return &GateGo{config: config}
}

func (gg GateGo) Run() error {
	table, err := NewHandlersTable(gg.config.Services)
	if err != nil {
		return err
	}

	// Gather checks and create checker
	checker := createChecker(gg.config.Services)
	checker.Start()

	// Blocking until exit or error
	err = gg.serveHandlers(table)
	if err != nil {
		return err
	}

	return nil
}

func createChecker(services []config.Service) Checker {
	checker := Checker{Delay: 5 * time.Second, OnFailure: func(err error) {}}

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

func (gg GateGo) serveHandlers(table HandlerTable) error {
	proxyMux := http.NewServeMux()

	proxyMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler := table.GetHandler(r.Host, r.URL.Path)

		if handler == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		handler.ServeHTTP(w, r)
	})

	addr := fmt.Sprintf("%s:%d", gg.config.Host, gg.config.Port)

	return serveMux(addr, proxyMux, gg.config.SSL.CertFile, gg.config.SSL.KeyFile)
}

func serveMux(addr string, mux *http.ServeMux, certfile *string, keyfile *string) error {
	supportTLS, err := checkTLSConfig(certfile, keyfile)
	if err != nil {
		return err
	}

	if supportTLS {
		log.Default().Printf("Serving proxy with TLS %s\n", addr)
		return http.ListenAndServeTLS(addr, *certfile, *keyfile, mux)
	} else {
		log.Default().Printf("Serving proxy %s\n", addr)
		return http.ListenAndServe(addr, mux)
	}
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
