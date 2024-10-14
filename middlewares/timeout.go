package middlewares

import (
	"context"
	"net/http"
	"time"
)

// NewTimeoutMiddleware returns an HTTP handler that wraps the provided handler with a timeout.
// If the processing takes longer than the specified timeout, it returns a 503 Service Unavailable error.
func NewTimeoutMiddleware(timeout time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a context with the specified timeout
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel() // Make sure to cancel the context when done

			// Create a new request with the timeout context
			r = r.WithContext(ctx)

			// Channel to capture when the request processing finishes
			done := make(chan struct{})

			go func() {
				// Serve the request
				next.ServeHTTP(w, r)
				// Signal that the request processing is done
				close(done)
			}()

			select {
			case <-ctx.Done():
				// If the context is canceled (due to timeout), return an error response
				if ctx.Err() == context.DeadlineExceeded {
					http.Error(w, "Request timed out", http.StatusGatewayTimeout)
				}
			case <-done:
				// If the request finished within the timeout, return the result
				return
			}
		})
	}
}
