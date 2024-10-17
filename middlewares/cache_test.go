package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCacheMiddleware(t *testing.T) {
	// Reset cache before each test
	responseCache.Flush()

	t.Run("Should not cache response with no cache headers", func(t *testing.T) {
		responseText := "test response"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte(responseText))
		})

		middleware := NewCacheMiddleware()(handler)
		req := httptest.NewRequest("GET", "/test", nil)

		// First request
		w1 := httptest.NewRecorder()
		middleware.ServeHTTP(w1, req)

		if w1.Body.String() != "test response" {
			t.Errorf("Expected 'test response', got '%s'", w1.Body.String())
		}

		responseText = "new response"

		// Second request - should be served from cache
		w2 := httptest.NewRecorder()
		middleware.ServeHTTP(w2, req)

		if w2.Body.String() != "new response" {
			t.Errorf("Expected not cached 'new response', got '%s'", w2.Body.String())
		}
	})

	t.Run("Should respect max-age Cache-Control header", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "max-age=1")
			w.WriteHeader(200)
			w.Write([]byte("cache-control test"))
		})

		middleware := NewCacheMiddleware()(handler)
		req := httptest.NewRequest("GET", "/cache-control", nil)

		// First request
		w1 := httptest.NewRecorder()
		middleware.ServeHTTP(w1, req)

		// Wait for less than max-age
		time.Sleep(time.Millisecond * 500)

		// Should still be cached
		w2 := httptest.NewRecorder()
		middleware.ServeHTTP(w2, req)

		if w2.Body.String() != "cache-control test" {
			t.Errorf("Expected cached response before max-age expiration")
		}

		// Wait for cache to expire
		time.Sleep(time.Millisecond * 1500)

		if _, found := responseCache.Get("/cache-control"); found {
			t.Error("Cache should have expired")
		}
	})

	t.Run("Should respect Expires header", func(t *testing.T) {
		responseText := "expires test"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expiresTime := time.Now().Add(2 * time.Second)
			w.Header().Set("Expires", expiresTime.Format(time.RFC1123))
			w.WriteHeader(200)
			w.Write([]byte(responseText))
		})

		middleware := NewCacheMiddleware()(handler)
		req := httptest.NewRequest("GET", "/expires", nil)

		// First request
		w1 := httptest.NewRecorder()
		middleware.ServeHTTP(w1, req)

		// Wait for less than expiration
		time.Sleep(time.Second * 1)

		responseText = "something else"

		// Should still be cached
		w2 := httptest.NewRecorder()
		middleware.ServeHTTP(w2, req)

		if w2.Body.String() != "expires test" {
			t.Errorf("Expected cached response before expiration")
		}

		// Wait for cache to expire
		time.Sleep(time.Second * 2)

		if _, found := responseCache.Get("/expires"); found {
			t.Error("Cache should have expired")
		}
	})

	t.Run("Should preserve response headers", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add expiration header
			expiresTime := time.Now().Add(50 * time.Second)
			w.Header().Set("Expires", expiresTime.Format(time.RFC1123))

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Custom-Header", "test-value")
			w.WriteHeader(200)
			w.Write([]byte(`{"message":"test"}`))
		})

		middleware := NewCacheMiddleware()(handler)
		req := httptest.NewRequest("GET", "/headers", nil)

		// First request
		w1 := httptest.NewRecorder()
		middleware.ServeHTTP(w1, req)

		// Second request - should preserve headers
		w2 := httptest.NewRecorder()
		middleware.ServeHTTP(w2, req)

		expectedHeaders := map[string]string{
			"Content-Type":    "application/json",
			"X-Custom-Header": "test-value",
		}

		for header, expectedValue := range expectedHeaders {
			if value := w2.Header().Get(header); value != expectedValue {
				t.Errorf("Expected header %s to be %s, got %s", header, expectedValue, value)
			}
		}
	})

	t.Run("Should handle invalid cache headers gracefully", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "max-age=invalid")
			w.Header().Set("Expires", "invalid-date")
			w.WriteHeader(200)
			w.Write([]byte("invalid headers test"))
		})

		middleware := NewCacheMiddleware()(handler)
		req := httptest.NewRequest("GET", "/invalid-headers", nil)

		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)

		if w.Body.String() != "invalid headers test" {
			t.Errorf("Expected normal response despite invalid headers")
		}
	})
}

func TestGetCacheMaxAge(t *testing.T) {
	tests := []struct {
		name         string
		cacheControl string
		expected     int
	}{
		{"Valid max-age", "max-age=60", 60},
		{"Multiple directives", "public, max-age=30", 30},
		{"Invalid max-age", "max-age=invalid", 0},
		{"No max-age", "public, private", 0},
		{"Empty string", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCacheMaxAge(tt.cacheControl)
			if result != tt.expected {
				t.Errorf("getCacheMaxAge(%s) = %d; want %d", tt.cacheControl, result, tt.expected)
			}
		})
	}
}

func TestGetCacheExpires(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name          string
		expiresHeader string
		wantZero      bool
	}{
		{"Valid date", now.Format(time.RFC1123), false},
		{"Invalid date", "invalid-date", true},
		{"Empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCacheExpires(tt.expiresHeader)
			if tt.wantZero && !result.IsZero() {
				t.Errorf("getCacheExpires(%s) expected zero time, got %v", tt.expiresHeader, result)
			}
			if !tt.wantZero && result.IsZero() {
				t.Errorf("getCacheExpires(%s) expected non-zero time, got zero time", tt.expiresHeader)
			}
		})
	}
}
