package middlewares_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/hvuhsg/gatego/middlewares"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to compare JSON
func assertJSONEqual(t *testing.T, expected, actual string) {
	var expectedMap, actualMap map[string]interface{}
	err := json.Unmarshal([]byte(expected), &expectedMap)
	require.NoError(t, err, "Error unmarshaling expected JSON")
	err = json.Unmarshal([]byte(actual), &actualMap)
	require.NoError(t, err, "Error unmarshaling actual JSON")
	assert.Equal(t, expectedMap, actualMap)
}

func TestOpenAPIValidationMiddleware(t *testing.T) {
	// Create a temporary OpenAPI spec file for testing
	specFile, err := os.CreateTemp("", "openapi-spec-*.yaml")
	require.NoError(t, err)
	defer os.Remove(specFile.Name())

	// Write a simple OpenAPI spec to the file
	specContent := `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      parameters:
        - name: param
          in: query
          required: true
          schema:
            type: string
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                required:
                  - message
                properties:
                  message:
                    type: string
                  status:
                    type: string
`
	_, err = specFile.Write([]byte(specContent))
	require.NoError(t, err)
	specFile.Close()

	// Create the middleware
	middleware, err := middlewares.NewOpenAPIValidationMiddleware(specFile.Name())
	require.NoError(t, err)

	tests := []struct {
		name           string
		url            string
		handler        http.HandlerFunc
		expectedStatus int
		expectedBody   string
		isJSON         bool
	}{
		{
			name: "Valid request and response",
			url:  "/test?param=value",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"message": "Hello, World!", "status": "ok"})
			}),
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Hello, World!","status":"ok"}`,
			isJSON:         true,
		},
		{
			name: "Valid request but invalid response (missing required field)",
			url:  "/test?param=value",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "error"})
			}),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Invalid response:",
			isJSON:         false,
		},
		{
			name: "Valid request but invalid response (wrong content type)",
			url:  "/test?param=value",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Hello, World!"))
			}),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Invalid response:",
			isJSON:         false,
		},
		{
			name: "Valid request but response with extra field",
			url:  "/test?param=value",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"message": "Hello, World!", "extra": "field"})
			}),
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Hello, World!","extra":"field"}`,
			isJSON:         true,
		},
		{
			name:           "Invalid path",
			url:            "/invalid",
			handler:        http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Error finding route:",
			isJSON:         false,
		},
		{
			name:           "Missing required parameter",
			url:            "/test",
			handler:        http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid request:",
			isJSON:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.url, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			wrappedHandler := middleware(tt.handler)
			wrappedHandler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.isJSON {
				assertJSONEqual(t, tt.expectedBody, rr.Body.String())
			} else {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestNewOpenAPIValidationMiddleware(t *testing.T) {
	tests := []struct {
		name        string
		specContent string
		expectError bool
	}{
		{
			name: "Valid OpenAPI spec",
			specContent: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
`,
			expectError: false,
		},
		{
			name:        "Invalid OpenAPI spec",
			specContent: "invalid: yaml: content",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			specFile, err := os.CreateTemp("", "openapi-spec-*.yaml")
			require.NoError(t, err)
			defer os.Remove(specFile.Name())

			_, err = specFile.Write([]byte(tt.specContent))
			require.NoError(t, err)
			specFile.Close()

			middleware, err := middlewares.NewOpenAPIValidationMiddleware(specFile.Name())

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, middleware)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, middleware)
			}
		})
	}
}
