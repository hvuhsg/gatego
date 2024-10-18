package middlewares

import (
	"net/http"
)

// OmitHeaders middleware removes specified headers from the response to enhance security.
func NewOmitHeadersMiddleware(headers []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			rc := NewRecorder()
			next.ServeHTTP(rc, r)

			// Omit headers from response
			for _, header := range headers {
				rc.Result().Header.Del(header)
			}

			rc.WriteHeadersTo(w)
			w.WriteHeader(rc.Result().StatusCode)
			w.Write(rc.Body.Bytes())
		})
	}
}
