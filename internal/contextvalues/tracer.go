package contextvalues

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// Define a custom type for context keys to avoid collisions
type tracerKeyType string

var tracerKey = tracerKeyType("tracer")

// Add tracer to context
func AddTracerToContext(ctx context.Context, tracer trace.Tracer) context.Context {
	return context.WithValue(ctx, tracerKey, tracer)
}

// Retrieve tracer from context
func TracerFromContext(ctx context.Context) trace.Tracer {
	var tracer trace.Tracer = nil
	if t, ok := ctx.Value(tracerKey).(trace.Tracer); ok {
		tracer = t
	}
	return tracer
}
