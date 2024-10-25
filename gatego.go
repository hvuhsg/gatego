package gatego

import (
	"context"
	"fmt"
	"time"

	"github.com/hvuhsg/gatego/internal/config"
	"github.com/hvuhsg/gatego/internal/contextvalues"
	"github.com/hvuhsg/gatego/pkg/monitor"
)

const serviceName = "gatego"

type GateGo struct {
	config  config.Config
	monitor *monitor.Monitor
	ctx     context.Context
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

	// Create checks start monitoring
	healthChecks := createMonitorChecks(gg.config.Services)
	gg.monitor = monitor.New(time.Second*5, healthChecks...)
	gg.monitor.Start()

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
		fmt.Println("\nShutting down...")
		return server.Shutdown(context.Background())
	}
}
