package middlewares

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

func NewAddHeadersMiddleware(headers map[string]string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := trace.SpanFromContext(r.Context())

			for header, value := range headers {
				r.Header.Set(header, value)
				span.AddEvent(fmt.Sprintf("Added header %s to request", header))
			}
			next.ServeHTTP(w, r)
		})
	}
}
