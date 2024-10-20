package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// MockSpanExporter is a test helper that captures exported spans
type MockSpanExporter struct {
	Spans []sdktrace.ReadOnlySpan
}

func (e *MockSpanExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	e.Spans = append(e.Spans, spans...)
	return nil
}

func (e *MockSpanExporter) Shutdown(ctx context.Context) error {
	return nil
}

func TestOpenTelemetryMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		config         OTELConfig
		handler        http.Handler
		request        *http.Request
		expectedStatus int
		validateSpan   func(*testing.T, sdktrace.ReadOnlySpan)
	}{
		{
			name: "successful request",
			config: OTELConfig{
				Endpoint:      "localhost:4317",
				SampleRatio:   1.0,
				ServiceDomain: "test.com",
				BasePath:      "/api",
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			request:        httptest.NewRequest(http.MethodGet, "/test", nil),
			expectedStatus: http.StatusOK,
			validateSpan: func(t *testing.T, span sdktrace.ReadOnlySpan) {
				assert.Equal(t, spanPrefix+"_request", span.Name())
				assert.Equal(t, trace.SpanKindServer, span.SpanKind())
				assert.Equal(t, codes.Ok, span.Status().Code)

				// Verify attributes
				hasAttribute := false
				for _, attr := range span.Attributes() {
					if attr.Key == "http.method" {
						hasAttribute = true
						assert.Equal(t, "GET", attr.Value.AsString())
					}
				}
				assert.True(t, hasAttribute, "span should have http.method attribute")
			},
		},
		{
			name: "error request",
			config: OTELConfig{
				Endpoint:      "localhost:4317",
				SampleRatio:   1.0,
				ServiceDomain: "test.com",
				BasePath:      "/api",
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			}),
			request:        httptest.NewRequest(http.MethodPost, "/test", nil),
			expectedStatus: http.StatusBadRequest,
			validateSpan: func(t *testing.T, span sdktrace.ReadOnlySpan) {
				assert.Equal(t, codes.Error, span.Status().Code)
				assert.Equal(t, http.StatusText(http.StatusBadRequest), span.Status().Description)

				// Verify status code attribute
				hasAttribute := false
				for _, attr := range span.Attributes() {
					if attr.Key == "http.status_code" {
						hasAttribute = true
						assert.Equal(t, int64(http.StatusBadRequest), attr.Value.AsInt64())
					}
				}
				assert.True(t, hasAttribute, "span should have http.status_code attribute")
			},
		},
		{
			name: "with sampling",
			config: OTELConfig{
				Endpoint:      "localhost:4317",
				SampleRatio:   0.5, // 50% sampling
				ServiceDomain: "test.com",
				BasePath:      "/api",
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			request:        httptest.NewRequest(http.MethodGet, "/test", nil),
			expectedStatus: http.StatusOK,
			validateSpan: func(t *testing.T, span sdktrace.ReadOnlySpan) {
				// Basic validation since sampling is probabilistic
				assert.NotEmpty(t, span.SpanContext().TraceID())
				assert.NotEmpty(t, span.SpanContext().SpanID())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock exporter
			mockExporter := &MockSpanExporter{}

			// Create a trace provider with the mock exporter
			tp := sdktrace.NewTracerProvider(
				sdktrace.WithBatcher(mockExporter),
				sdktrace.WithSampler(sdktrace.TraceIDRatioBased(tt.config.SampleRatio)),
			)

			// Set the trace provider
			otel.SetTracerProvider(tp)

			// Create the middleware
			middleware, err := NewOpenTelemetryMiddleware(context.Background(), tt.config)
			require.NoError(t, err)

			// Create a test server with the middleware
			handler := middleware(tt.handler)
			server := httptest.NewServer(handler)
			defer server.Close()

			// Send the request
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, tt.request)

			// Verify response
			assert.Equal(t, tt.expectedStatus, recorder.Code)

			// Shutdown the provider to flush spans
			require.NoError(t, tp.Shutdown(context.Background()))

			// Verify spans
			if len(mockExporter.Spans) > 0 {
				tt.validateSpan(t, mockExporter.Spans[0])
			}
		})
	}
}
