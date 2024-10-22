package gatego

import (
	"context"
	"time"

	"github.com/hvuhsg/gatego/config"
	"github.com/hvuhsg/gatego/contextvalues"
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

	server, err := newServer(gg.ctx, gg.config, useOtel)
	if err != nil {
		return err
	}
	defer server.Shutdown(gg.ctx)

	serveErrChan, err := server.serve(gg.config.TLS.CertFile, gg.config.TLS.KeyFile)
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
	checker := &Checker{Delay: 5 * time.Second}

	for _, service := range services {
		for _, path := range service.Paths {
			for _, checkConfig := range path.Checks {
				check := Check{
					Name:      checkConfig.Name,
					Cron:      checkConfig.Cron,
					URL:       checkConfig.URL,
					Method:    checkConfig.Method,
					Timeout:   checkConfig.Timeout,
					Headers:   checkConfig.Headers,
					OnFailure: checkConfig.OnFailure,
				}

				checker.Checks = append(checker.Checks, check)
			}
		}
	}

	return checker
}
