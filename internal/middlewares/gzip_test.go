package middlewares_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hvuhsg/gatego/internal/middlewares"
)

// Helper to decode gzip data
func decodeGzip(t *testing.T, gzippedBody []byte) string {
	gzipReader, err := gzip.NewReader(bytes.NewReader(gzippedBody))
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gzipReader.Close()

	var decodedBody bytes.Buffer
	if _, err := io.Copy(&decodedBody, gzipReader); err != nil {
		t.Fatalf("failed to decode gzip body: %v", err)
	}

	return decodedBody.String()
}

// TestGzipMiddleware_NoGzipSupport tests the middleware when the client does not support gzip
func TestGzipMiddleware_NoGzipSupport(t *testing.T) {
	// Create a test handler to be wrapped by the middleware
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World"))
	})

	// Wrap the handler with GzipMiddleware
	handler := middlewares.GzipMiddleware(nextHandler)

	// Create a new HTTP request without gzip support
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "identity")

	// Record the response
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check that the response is not gzipped
	if encoding := rr.Header().Get("Content-Encoding"); encoding != "" {
		t.Errorf("expected no gzip encoding, got %s", encoding)
	}

	// Check the body content
	if rr.Body.String() != "Hello, World" {
		t.Errorf("expected 'Hello, World', got %s", rr.Body.String())
	}
}

// TestGzipMiddleware_WithGzipSupport tests the middleware when the client supports gzip
func TestGzipMiddleware_WithGzipSupport(t *testing.T) {
	// Create a test handler to be wrapped by the middleware
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World"))
	})

	// Wrap the handler with GzipMiddleware
	handler := middlewares.GzipMiddleware(nextHandler)

	// Create a new HTTP request with gzip support
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	// Record the response
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check that the response is gzipped
	if encoding := rr.Header().Get("Content-Encoding"); encoding != "gzip" {
		t.Errorf("expected gzip encoding, got %s", encoding)
	}

	// Decode the gzipped response body
	gzippedBody := rr.Body.Bytes()
	decodedBody := decodeGzip(t, gzippedBody)

	// Check the body content
	if decodedBody != "Hello, World" {
		t.Errorf("expected 'Hello, World', got %s", decodedBody)
	}
}

// TestGzipMiddleware_StatusCode tests that the middleware preserves status codes
func TestGzipMiddleware_StatusCode(t *testing.T) {
	// Create a test handler to be wrapped by the middleware
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Created"))
	})

	// Wrap the handler with GzipMiddleware
	handler := middlewares.GzipMiddleware(nextHandler)

	// Create a new HTTP request with gzip support
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	// Record the response
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check that the response is gzipped
	if encoding := rr.Header().Get("Content-Encoding"); encoding != "gzip" {
		t.Errorf("expected gzip encoding, got %s", encoding)
	}

	// Check that the status code is preserved
	if status := rr.Result().StatusCode; status != http.StatusCreated {
		t.Errorf("expected status code %d, got %d", http.StatusCreated, status)
	}

	// Decode the gzipped response body
	gzippedBody := rr.Body.Bytes()
	decodedBody := decodeGzip(t, gzippedBody)

	// Check the body content
	if decodedBody != "Created" {
		t.Errorf("expected 'Created', got %s", decodedBody)
	}
}
