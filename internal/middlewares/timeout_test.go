package middlewares_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hvuhsg/gatego/internal/middlewares"
)

func TestTimeoutMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		timeout        time.Duration
		handlerSleep   time.Duration
		expectedStatus int
	}{
		{
			name:           "Request completes before timeout",
			timeout:        100 * time.Millisecond,
			handlerSleep:   50 * time.Millisecond,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Request times out",
			timeout:        50 * time.Millisecond,
			handlerSleep:   100 * time.Millisecond,
			expectedStatus: http.StatusGatewayTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler that sleeps for the specified duration
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(tt.handlerSleep)
				w.WriteHeader(http.StatusOK)
			})

			// Wrap the handler with our middleware
			wrappedHandler := middlewares.NewTimeoutMiddleware(tt.timeout)(handler)

			// Create a test request
			req, err := http.NewRequest("GET", "/test", nil)
			if err != nil {
				t.Fatal(err)
			}

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Serve the request using our wrapped handler
			wrappedHandler.ServeHTTP(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}

func TestTimeoutMiddlewareCancelContext(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep until the context is canceled
		<-r.Context().Done()
		// Check if the context was canceled due to a timeout
		if r.Context().Err() == context.DeadlineExceeded {
			w.WriteHeader(http.StatusGatewayTimeout)
		}

		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middlewares.NewTimeoutMiddleware(50 * time.Millisecond)(handler)

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusGatewayTimeout {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusGatewayTimeout)
	}
}
