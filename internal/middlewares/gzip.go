package middlewares

import (
	"compress/gzip"
	"net/http"
	"strings"
)

// GzipMiddleware compresses the response using gzip if the client supports it
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the client accepts gzip encoding
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// Client doesn't support gzip, serve the next handler
			next.ServeHTTP(w, r)
			return
		}

		// Create a gzip.Writer
		gzipWriter := gzip.NewWriter(w)
		defer gzipWriter.Close()

		// Serve the next handler, writing the response into the ResponseCapture
		rc := NewRecorder()
		next.ServeHTTP(rc, r)

		rc.WriteHeadersTo(w)

		w.Header().Del("Content-Length")
		w.Header().Set("Content-Encoding", "gzip") // Set Content-Encoding header

		w.WriteHeader(rc.Result().StatusCode)

		gzipWriter.Write(rc.Body.Bytes())
	})
}
