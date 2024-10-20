package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"
)

const (
	serviceName    = "gatego"
	spanPrefix     = "gatego"
	defaultTimeout = 5 * time.Second
)

type OTELConfig struct {
	ServiceVersion string
	Endpoint       string // OTLP gRPC endpoint
	SampleRatio    float64
	Credentials    credentials.TransportCredentials
	ServiceDomain  string
	BasePath       string
}

func NewOpenTelemetryMiddleware(ctx context.Context, config OTELConfig) (Middleware, error) {
	// Connection security option
	securityOpt := otlptracegrpc.WithInsecure()
	if config.Credentials != nil {
		securityOpt = otlptracegrpc.WithTLSCredentials(config.Credentials) // TODO: allow creds
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(config.Endpoint), // OTLP gRPC endpoint
			securityOpt,
			otlptracegrpc.WithTimeout(defaultTimeout),
		),
	)
	if err != nil {
		return nil, err
	}

	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
		semconv.TelemetrySDKLanguageGo,
		attribute.String("version", config.ServiceVersion),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(float64(config.SampleRatio))), // Just for testing
	)
	otel.SetTracerProvider(tp)

	tracer := tp.Tracer(serviceName)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create span for request
			spanName := fmt.Sprintf("%s_%s_%s", spanPrefix, r.Method, r.URL.Path)
			ctx, span := tracer.Start(
				r.Context(),
				spanName,
				trace.WithAttributes(semconv.NetAttributesFromHTTPRequest("", r)...),
				trace.WithSpanKind(trace.SpanKindServer),
			)
			defer span.End()

			// Add request-specific attributes
			span.SetAttributes(
				semconv.HTTPUserAgentKey.String(r.UserAgent()),
			)
			span.SetAttributes(
				semconv.HTTPServerAttributesFromHTTPRequest(config.ServiceDomain, config.BasePath, r)...,
			)

			// Handle panic recovery
			defer func() {
				if err := recover(); err != nil {
					span.SetStatus(codes.Error, fmt.Sprintf("panic: %v", err))
					span.RecordError(fmt.Errorf("%v", err))
					panic(err) // Re-panic after recording error
				}
			}()

			// Add span to request context
			rc := NewRecorder()
			next.ServeHTTP(rc, r.WithContext(ctx))

			// Set status and attributes based on response code
			statusCode := rc.Result().StatusCode
			span.SetAttributes(semconv.HTTPAttributesFromHTTPStatusCode(statusCode)...)
			if statusCode >= 400 {
				span.SetStatus(codes.Error, http.StatusText(statusCode))
				if statusCode >= 500 {
					span.RecordError(fmt.Errorf("server error: %d", statusCode))
				}
			} else {
				span.SetStatus(codes.Ok, "")
			}

			// Add response information
			span.SetAttributes(
				attribute.Int64("http.response_size", rc.Result().ContentLength),
				attribute.String("http.response_content_type", rc.Result().Header.Get("Content-Type")),
			)

			// Add trace context to response headers
			w.Header().Set("X-Trace-ID", span.SpanContext().TraceID().String())

			// Return response
			rc.WriteTo(w)
		})
	}, nil
}
