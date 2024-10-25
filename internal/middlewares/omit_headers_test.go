package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hvuhsg/gatego/internal/middlewares"
)

// TestOmitHeadersMiddleware_OmitResponseHeaders tests that headers are omitted from the response
func TestOmitHeadersMiddleware_OmitResponseHeaders(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Authorization", "Bearer some-secret-token")
		w.Header().Set("X-API-Key", "secret-api-key")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	headers := []string{"Authorization", "X-API-Key"}
	handler := middlewares.NewOmitHeadersMiddleware(headers)(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Authorization") != "" {
		t.Errorf("expected 'Authorization' header to be omitted, got %s", rr.Header().Get("Authorization"))
	}
	if rr.Header().Get("X-API-Key") != "" {
		t.Errorf("expected 'X-API-Key' header to be omitted, got %s", rr.Header().Get("X-API-Key"))
	}

	if rr.Body.String() != "OK" {
		t.Errorf("expected 'OK', got %s", rr.Body.String())
	}
}
