package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveBaseURLPath(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		fullPath string
		want     string
		wantErr  bool
	}{
		{
			name:     "simple path",
			basePath: "/api",
			fullPath: "/api/file.txt",
			want:     "/file.txt",
			wantErr:  false,
		},
		{
			name:     "path with multiple segments",
			basePath: "/api/v1",
			fullPath: "/api/v1/docs/file.txt",
			want:     "/docs/file.txt",
			wantErr:  false,
		},
		{
			name:     "paths with trailing slashes",
			basePath: "/api/",
			fullPath: "/api/file.txt/",
			want:     "/file.txt",
			wantErr:  false,
		},
		{
			name:     "paths without leading slashes",
			basePath: "api",
			fullPath: "api/file.txt",
			want:     "/file.txt",
			wantErr:  false,
		},
		{
			name:     "path not in base path",
			basePath: "/api",
			fullPath: "/other/file.txt",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "empty paths",
			basePath: "",
			fullPath: "/file.txt",
			want:     "/file.txt",
			wantErr:  false,
		},
		{
			name:     "identical paths",
			basePath: "/api",
			fullPath: "/api",
			want:     "/",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := removeBaseURLPath(tt.basePath, tt.fullPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("removeBaseURLPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("removeBaseURLPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFiles_ServeHTTP(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "files_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file
	testContent := []byte("test file content")
	testFilePath := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFilePath, testContent, 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name           string
		basePath       string
		requestPath    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid file request",
			basePath:       "/files",
			requestPath:    "/files/test.txt",
			expectedStatus: http.StatusOK,
			expectedBody:   "test file content",
		},
		{
			name:           "file not found",
			basePath:       "/files",
			requestPath:    "/files/nonexistent.txt",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
		{
			name:           "path outside base path",
			basePath:       "/files",
			requestPath:    "/other/test.txt",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new Files handler
			files := NewFiles(tmpDir, tt.basePath)

			// Create a test request
			req := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)
			w := httptest.NewRecorder()

			// Serve the request
			files.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("ServeHTTP() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			// Check response body
			if w.Body.String() != tt.expectedBody {
				t.Errorf("ServeHTTP() body = %v, want %v", w.Body.String(), tt.expectedBody)
			}
		})
	}
}
