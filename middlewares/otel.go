package middlewares

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hvuhsg/gatego/contextvalues"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "request"
const spanName = "middlewares"

type OTELConfig struct {
	ServiceDomain string
	BasePath      string
}

func NewOpenTelemetryMiddleware(ctx context.Context, config OTELConfig) (Middleware, error) {
	tp := otel.GetTracerProvider()
	tracer := tp.Tracer(tracerName)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add tracer to request context
			r = r.WithContext(contextvalues.AddTracerToContext(r.Context(), tracer))

			// Create span for request
			ctx, span := tracer.Start(
				r.Context(),
				spanName,
				trace.WithAttributes(semconv.NetAttributesFromHTTPRequest("", r)...),
				trace.WithSpanKind(trace.SpanKindServer),
			)
			defer span.End()

			// Add request-specific attributes
			attrs := make([]attribute.KeyValue, 0)
			attrs = append(attrs, semconv.HTTPUserAgentKey.String(r.UserAgent()))
			attrs = append(attrs, semconv.HTTPServerAttributesFromHTTPRequest(config.ServiceDomain, config.BasePath, r)...)
			span.SetAttributes(
				attrs...,
			)

			// Handle panic recovery
			defer func() {
				if err := recover(); err != nil {
					span.SetStatus(codes.Error, fmt.Sprintf("panic: %v", err))
					span.RecordError(fmt.Errorf("%v", err))
					panic(err) // Re-panic after recording error
				}
			}()

			// Propegate open telemetry context via the request to the upstream service
			otel.GetTextMapPropagator().Inject(r.Context(), propagation.HeaderCarrier(r.Header))

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

			// Return response
			rc.WriteTo(w)
		})
	}, nil
}
