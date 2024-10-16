package middlewares

import (
	"net/http"
)

// OmitHeaders middleware removes specified headers from the response to enhance security.
func NewOmitHeadersMiddleware(headers []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			rc := NewResponseCapture(w)
			next.ServeHTTP(rc, r)

			// Omit headers from response
			for _, header := range headers {
				rc.Header().Del(header)
			}

			w.WriteHeader(rc.status)
			w.Write(rc.buffer.Bytes())
		})
	}
}
