package middlewares

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

// OmitHeaders middleware removes specified headers from the response to enhance security.
func NewOmitHeadersMiddleware(headersToOmit []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := trace.SpanFromContext(r.Context())

			rc := NewRecorder()
			next.ServeHTTP(rc, r)

			// Omit headers from response
			for _, header := range headersToOmit {
				if rc.Result().Header.Get(header) != "" {
					rc.Result().Header.Del(header)
					span.AddEvent(fmt.Sprintf("Removed response header %s", header))
				}
			}

			rc.WriteHeadersTo(w)
			w.WriteHeader(rc.Result().StatusCode)
			w.Write(rc.Body.Bytes())
		})
	}
}
