package middlewares

import "net/http"

func NewAddHeadersMiddleware(headers map[string]string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for header, value := range headers {
				r.Header.Set(header, value)
			}
			next.ServeHTTP(w, r)
		})
	}
}
