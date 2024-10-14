package middlewares

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestLoggingMiddleware(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		statusCode       int
		responseBody     string
		expectedLogParts []string
	}{
		{
			name:         "GET request",
			method:       "GET",
			path:         "/test",
			statusCode:   200,
			responseBody: "OK",
			expectedLogParts: []string{
				"GET",
				"/test",
				"HTTP/1.1",
				"200",
				"2",
				"http://example.com/test",
			},
		},
		{
			name:         "POST request with 404",
			method:       "POST",
			path:         "/notfound",
			statusCode:   404,
			responseBody: "Not Found",
			expectedLogParts: []string{
				"POST",
				"/notfound",
				"HTTP/1.1",
				"404",
				"9",
				"http://example.com/notfound",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture the log output
			buf := &bytes.Buffer{}

			// Create a test handler that returns the specified status code and body
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			})

			// Create the logging middleware
			loggingMiddleware := NewLoggingMiddleware(buf)

			// Create a test server with the logging middleware
			ts := httptest.NewServer(loggingMiddleware(testHandler))
			defer ts.Close()

			// Create and send the request
			req, _ := http.NewRequest(tt.method, ts.URL+tt.path, nil)
			req.Host = "example.com" // Set a consistent host for testing
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Error making request: %v", err)
			}
			defer resp.Body.Close()

			// Check the response
			if resp.StatusCode != tt.statusCode {
				t.Errorf("Expected status code %d, got %d", tt.statusCode, resp.StatusCode)
			}

			// Check the log output
			logOutput := buf.String()
			for _, expectedPart := range tt.expectedLogParts {
				if !strings.Contains(logOutput, expectedPart) {
					t.Errorf("Expected log to contain '%s', but it didn't. Log: %s", expectedPart, logOutput)
				}
			}

			// Check for the presence of a timestamp in the expected format
			timeStampFormat := "[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}"
			if matched, _ := regexp.MatchString(timeStampFormat, logOutput); !matched {
				t.Errorf("Expected log to contain a timestamp in format YYYY-MM-DD HH:MM:SS, but it didn't. Log: %s", logOutput)
			}

			// Check for the presence of a duration
			if !strings.Contains(logOutput, "ms") && !strings.Contains(logOutput, "s") {
				t.Errorf("Expected log to contain a duration, but it didn't. Log: %s", logOutput)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration int64
		expected string
	}{
		{"Less than a second", 500, "500ms"},
		{"Exactly one second", 1000, "1.0s"},
		{"More than a second", 1500, "1.5s"},
		{"Multiple seconds", 3750, "3.8s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("formatDuration(%d) = %s; want %s", tt.duration, result, tt.expected)
			}
		})
	}
}
