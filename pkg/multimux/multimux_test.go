package multimux

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterHandler(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		pattern string
	}{
		{"basic registration", "example.com", "/path"},
		{"with port", "example.com:8080", "/path"},
		{"uppercase host", "EXAMPLE.COM", "/path"},
		{"uppercase pattern", "/PATH", "/path"},
		{"with subdomain", "sub.example.com", "/path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mm := NewMultiMux()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			mm.RegisterHandler(tt.host, tt.pattern, handler)

			cleanedHost := cleanHost(tt.host)
			mux, exists := mm.Hosts.Load(cleanedHost)
			if !exists {
				t.Errorf("Host %s was not registered", cleanedHost)
			}
			if mux == nil {
				t.Errorf("ServeMux for host %s is nil", cleanedHost)
			}
		})
	}
}

func TestServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		host           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "existing host and path",
			host:           "example.com",
			path:           "/test",
			expectedStatus: http.StatusOK,
			expectedBody:   "handler1",
		},
		{
			name:           "existing host with port",
			host:           "example.com:8080",
			path:           "/test",
			expectedStatus: http.StatusOK,
			expectedBody:   "handler1",
		},
		{
			name:           "non-existing host",
			host:           "unknown.com",
			path:           "/test",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mm := NewMultiMux()

			// Register a test handler
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "handler1")
			})
			mm.RegisterHandler("example.com", "/test", handler)

			// Create test request
			req := httptest.NewRequest("GET", "http://"+tt.host+tt.path, nil)
			req.Host = tt.host
			w := httptest.NewRecorder()

			// Serve the request
			mm.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check response body if expected
			if tt.expectedBody != "" && w.Body.String() != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestCleanHost(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"example.com", "example.com"},
		{"EXAMPLE.COM", "example.com"},
		{"example.com:8080", "example.com"},
		{"EXAMPLE.COM:8080", "example.com"},
		{"sub.example.com:8080", "sub.example.com"},
		{"localhost", "localhost"},
		{"localhost:8080", "localhost"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := cleanHost(tt.input)
			if result != tt.expected {
				t.Errorf("cleanHost(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRemovePort(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"example.com", "example.com"},
		{"example.com:8080", "example.com"},
		{"example.com:80", "example.com"},
		{"localhost:8080", "localhost"},
		{"127.0.0.1:8080", "127.0.0.1"},
		{"[::1]:8080", "[::1]"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := removePort(tt.input)
			if result != tt.expected {
				t.Errorf("removePort(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}
