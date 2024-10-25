package middlewares_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hvuhsg/gatego/internal/middlewares"
)

// TestRequestSizeLimitMiddleware tests the RequestSizeLimitMiddleware
func TestRequestSizeLimitMiddleware(t *testing.T) {
	tests := []struct {
		name         string
		body         []byte
		maxSize      uint64
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Within limit",
			body:         []byte("This is within the limit."),
			maxSize:      30,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Exactly at limit",
			body:         bytes.Repeat([]byte("A"), 30), // 30 bytes
			maxSize:      30,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Exceeds limit",
			body:         bytes.Repeat([]byte("A"), 31), // 31 bytes
			maxSize:      30,
			expectedCode: http.StatusRequestEntityTooLarge,
			expectedBody: "Request body too large. Maximum allowed size is 30 bytes.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request with the test body
			req := httptest.NewRequest("POST", "http://example.com", bytes.NewReader(tt.body))
			rr := httptest.NewRecorder()

			// Use a simple handler that just returns OK
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Create the middleware with the specified maxSize
			middleware := middlewares.NewRequestSizeLimitMiddleware(tt.maxSize)

			// Serve the request through the middleware
			middleware(handler).ServeHTTP(rr, req)

			// Check the response code
			if rr.Code != tt.expectedCode {
				t.Errorf("expected status code %d, got %d", tt.expectedCode, rr.Code)
			}

			// Check the response body if applicable
			if tt.expectedBody != "" {
				if rr.Body.String() != tt.expectedBody {
					t.Errorf("expected response body %q, got %q", tt.expectedBody, rr.Body.String())
				}
			}
		})
	}
}
