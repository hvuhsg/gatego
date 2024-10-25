package gatego

import (
	"context"
	"errors"
	"time"

	"github.com/hvuhsg/gatego/internal/contextvalues"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type otelConfig struct {
	TraceCollectorEndpoint  string
	MetricCollectorEndpoint string
	LogsCollectorEndpoint   string
	CollectorTimeout        time.Duration
	ServiceName             string
	SampleRatio             float64
}

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func setupOTelSDK(ctx context.Context, conf otelConfig) (func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) (func(context.Context) error, error) {
		err := errors.Join(inErr, shutdown(ctx))
		return nil, err
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up resource
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(conf.ServiceName),
		semconv.TelemetrySDKLanguageGo,
		attribute.String("version", contextvalues.VersionFromContext(ctx)),
	)

	// Set up trace provider.
	tracerProvider, err := newTraceProvider(ctx, resource, conf.TraceCollectorEndpoint, conf.CollectorTimeout, conf.SampleRatio)
	if err != nil {
		return handleErr(err)
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, err := newMeterProvider(ctx, resource, conf.TraceCollectorEndpoint, conf.CollectorTimeout)
	if err != nil {
		return handleErr(err)
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Set up logger provider.
	loggerProvider, err := newLoggerProvider(ctx, resource, conf.TraceCollectorEndpoint, conf.CollectorTimeout)
	if err != nil {
		return handleErr(err)
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	return shutdown, err
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(ctx context.Context, resource *resource.Resource, endpoint string, timeout time.Duration, sampleRatio float64) (*trace.TracerProvider, error) {
	exporter, err := otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(endpoint), // OTLP gRPC endpoint
			otlptracegrpc.WithTimeout(timeout),
			otlptracegrpc.WithInsecure(),
		),
	)
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithResource(resource),
		trace.WithBatcher(exporter),
		trace.WithSampler(trace.TraceIDRatioBased(sampleRatio)),
	)
	return traceProvider, nil
}

func newMeterProvider(ctx context.Context, resource *resource.Resource, endpoint string, timeout time.Duration) (*metric.MeterProvider, error) {
	exporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint(endpoint), // OTLP gRPC endpoint
		otlpmetricgrpc.WithTimeout(timeout),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(resource),
		metric.WithReader(
			metric.NewPeriodicReader(
				exporter,
			),
		),
	)

	return meterProvider, nil
}

func newLoggerProvider(ctx context.Context, resource *resource.Resource, endpoint string, timeout time.Duration) (*log.LoggerProvider, error) {
	exporter, err := otlploggrpc.New(
		ctx,
		otlploggrpc.WithEndpoint(endpoint), // OTLP gRPC endpoint
		otlploggrpc.WithTimeout(timeout),
		otlploggrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithResource(resource),
		log.WithProcessor(log.NewBatchProcessor(exporter)),
	)
	return loggerProvider, nil
}
