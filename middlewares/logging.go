package middlewares

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func formatDuration(ms int64) string {
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.1fs", float64(ms)/1000)
}

// Logging middleware log the request / response with the log style of nginx
func NewLoggingMiddleware(out io.Writer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now().UnixMilli()

			rh := &responseHook{ResponseWriter: w, respSize: 0}
			next.ServeHTTP(rh, r)

			end := time.Now().UnixMilli()

			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			fullURL := fmt.Sprintf("%s://%s%s", scheme, r.Host, r.URL.String())

			method := r.Method
			path := r.URL.Path
			responseSize := rh.respSize
			remoteAddr := r.RemoteAddr
			date := time.Now().Format("2006-01-02 15:04:05")
			userAgent := r.UserAgent()
			statusCode := rh.statusCode
			duration := formatDuration(end - start)

			fmt.Fprintf(out, "%s - - [%s] \"%s %s %s\" %d %d %s \"%s\" \"%s\"\n", remoteAddr, date, method, path, r.Proto, statusCode, responseSize, duration, fullURL, userAgent)
		})
	}
}

type responseHook struct {
	http.ResponseWriter
	respSize   int
	statusCode int
}

func (rh *responseHook) Write(b []byte) (int, error) {
	// Save the length of the response
	rh.respSize += len(b)

	return rh.ResponseWriter.Write(b)
}

func (rh *responseHook) WriteHeader(statusCode int) {
	// Save status code
	rh.statusCode = statusCode

	rh.ResponseWriter.WriteHeader(statusCode)
}
