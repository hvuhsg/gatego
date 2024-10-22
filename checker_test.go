package gatego

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheck_run(t *testing.T) {
	tests := []struct {
		name           string
		server         *httptest.Server
		check          Check
		expectedError  bool
		serverResponse int
	}{
		{
			name: "successful check",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})),
			check: Check{
				Name:    "test-check",
				Method:  "GET",
				Timeout: 5 * time.Second,
				Headers: map[string]string{"X-Test": "test-value"},
			},
			expectedError:  false,
			serverResponse: http.StatusOK,
		},
		{
			name: "failed check - wrong status code",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			})),
			check: Check{
				Name:    "test-check-fail",
				Method:  "GET",
				Timeout: 5 * time.Second,
			},
			expectedError:  true,
			serverResponse: http.StatusInternalServerError,
		},
		{
			name: "check with timeout",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(2 * time.Second)
				w.WriteHeader(http.StatusOK)
			})),
			check: Check{
				Name:    "test-check-timeout",
				Method:  "GET",
				Timeout: 1 * time.Second,
			},
			expectedError: true,
		},
		{
			name: "check with custom headers",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("X-Custom") != "custom-value" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
			})),
			check: Check{
				Name:    "test-check-headers",
				Method:  "GET",
				Timeout: 5 * time.Second,
				Headers: map[string]string{"X-Custom": "custom-value"},
			},
			expectedError:  false,
			serverResponse: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.server.Close()

			tt.check.URL = tt.server.URL
			tt.check.run(func(error) {})
		})
	}
}

func TestChecker_Start(t *testing.T) {
	tests := []struct {
		name          string
		checker       Checker
		expectedError bool
	}{
		{
			name: "successful start",
			checker: Checker{
				Delay: 1 * time.Second,
				Checks: []Check{
					{
						Name:    "test-check",
						Cron:    "* * * * *",
						Method:  "GET",
						URL:     "http://example.com",
						Timeout: 5 * time.Second,
					},
				},
			},
			expectedError: false,
		},
		{
			name: "invalid cron expression",
			checker: Checker{
				Delay: 1 * time.Second,
				Checks: []Check{
					{
						Name:    "test-check-invalid-cron",
						Cron:    "invalid",
						Method:  "GET",
						URL:     "http://example.com",
						Timeout: 5 * time.Second,
					},
				},
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.checker.Start()
			if (err != nil) != tt.expectedError {
				t.Errorf("Checker.Start() error = %v, expectedError %v", err, tt.expectedError)
			}

			// Clean up scheduler if it was created
			if tt.checker.scheduler != nil {
				tt.checker.scheduler.Stop()
			}
		})
	}
}

func TestChecker_OnFailure(t *testing.T) {
	tests := []struct {
		name          string
		checker       Checker
		expectedError bool
	}{
		{
			name: "on failure command with valid command",
			checker: Checker{
				Delay: 1 * time.Second,
				Checks: []Check{
					{
						Name:      "test-check-failure",
						Cron:      "* * * * *",
						Method:    "GET",
						URL:       "http://example.com",
						Timeout:   5 * time.Second,
						OnFailure: "echo check '$check_name' failed at $date: $error",
					},
				},
			},
			expectedError: false,
		},
		{
			name: "on failure command with invalid command",
			checker: Checker{
				Delay: 1 * time.Second,
				Checks: []Check{
					{
						Name:      "test-check-failure",
						Cron:      "* * * * *",
						Method:    "GET",
						URL:       "http://example.com",
						Timeout:   5 * time.Second,
						OnFailure: "invalidCommand $error",
					},
				},
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate a failure scenario by injecting an error
			err := errors.New("Connection timeout")
			err = handleFailure(tt.checker.Checks[0], err)

			// Check if an error was returned and if it matches the expected result
			if (err != nil) != tt.expectedError {
				t.Errorf("handleFailure() error = %v, expectedError %v", err, tt.expectedError)
			}

			// Clean up scheduler if it was created
			if tt.checker.scheduler != nil {
				tt.checker.scheduler.Stop()
			}
		})
	}
}

// TestCheckWithMockServer tests the Check struct with a mock HTTP server
func TestCheckWithMockServer(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method
		if r.Method != http.MethodGet {
			t.Errorf("Expected method %s, got %s", http.MethodGet, r.Method)
		}

		// Verify headers
		if r.Header.Get("X-Test") != "test-value" {
			t.Errorf("Expected header X-Test: test-value, got %s", r.Header.Get("X-Test"))
		}

		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	check := Check{
		Name:    "test-check",
		Method:  http.MethodGet,
		URL:     server.URL,
		Timeout: 5 * time.Second,
		Headers: map[string]string{"X-Test": "test-value"},
	}

	check.run(func(error) {})
}
