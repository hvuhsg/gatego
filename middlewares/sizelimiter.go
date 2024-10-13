package middlewares

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// NewRequestSizeLimitMiddleware limits the size of the request body to the specified limit in bytes.
func NewRequestSizeLimitMiddleware(maxSize uint64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a buffer to read the body
			buf := new(bytes.Buffer)

			// Use io.LimitReader to limit the size of the request body
			limitedReader := io.LimitReader(r.Body, int64(maxSize+1)) // Allow one extra byte for overflow detection
			_, err := io.Copy(buf, limitedReader)                     // Copy the limited input into the buffer

			// Check for errors
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusInternalServerError)
				return
			}

			// Check if we exceeded the maximum size
			if buf.Len() > int(maxSize) {
				http.Error(w, fmt.Sprintf("Request body too large. Maximum allowed size is %d bytes.", maxSize), http.StatusRequestEntityTooLarge)
				return
			}

			// Restore the request body for further processing
			r.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))

			// Proceed to the next handler
			next.ServeHTTP(w, r)
		})
	}
}
