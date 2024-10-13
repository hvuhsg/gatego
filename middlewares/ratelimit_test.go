package middlewares_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hvuhsg/gatego/middlewares"
)

func TestRateLimitExceeded(t *testing.T) {
	limits := []string{"ip-1/s"}
	rateLimitMiddleware, err := middlewares.NewRateLimitMiddleware(limits)
	if err != nil {
		t.Fatalf("Error creating middleware: %v", err)
	}

	handler := rateLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create a test server
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()

	// First request should pass
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", rr.Code)
	}

	// Second request should fail (rate limit exceeded)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("expected status TooManyRequests, got %v", rr.Code)
	}
}

func TestRateLimitHeaders(t *testing.T) {
	limits := []string{"ip-1/s"}
	rateLimitMiddleware, err := middlewares.NewRateLimitMiddleware(limits)
	if err != nil {
		t.Fatalf("Error creating middleware: %v", err)
	}

	handler := rateLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()

	// First request should pass
	handler.ServeHTTP(rr, req)

	// Second request should fail
	handler.ServeHTTP(rr, req)

	if rr.Header().Get("X-RateLimit-Limit") != "1" {
		t.Errorf("expected X-RateLimit-Limit 1, got %s", rr.Header().Get("X-RateLimit-Limit"))
	}
	if rr.Header().Get("X-RateLimit-Remaining") != "0" {
		t.Errorf("expected X-RateLimit-Remaining 0, got %s", rr.Header().Get("X-RateLimit-Remaining"))
	}
}

func TestInvalidRateLimitConfig(t *testing.T) {
	limits := []string{"ip-invalid/s"}
	_, err := middlewares.NewRateLimitMiddleware(limits)
	if err == nil {
		t.Errorf("expected error for invalid config, got none")
	}
}

func TestUnsupportedZone(t *testing.T) {
	limits := []string{"unsupported-10/s"}
	_, err := middlewares.NewRateLimitMiddleware(limits)
	if !errors.Is(err, middlewares.ErrZoneNotSupported) {
		t.Errorf("expected ErrZoneNotSupported, got %v", err)
	}
}

func TestRateLimitConcurrentRequests(t *testing.T) {
	limits := []string{"ip-3/s"}
	rateLimitMiddleware, err := middlewares.NewRateLimitMiddleware(limits)
	if err != nil {
		t.Fatalf("Error creating middleware: %v", err)
	}

	handler := rateLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	var wg sync.WaitGroup
	var rateLimitedCount int
	var mu sync.Mutex

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			mu.Lock()
			if rr.Code == http.StatusTooManyRequests {
				rateLimitedCount++
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	if rateLimitedCount != 2 {
		t.Errorf("expected 2 rate limited requests, got %d", rateLimitedCount)
	}
}

func TestRateLimitDifferentTimeWindows(t *testing.T) {
	limits := []string{"ip-2/s", "ip-5/m"}
	rateLimitMiddleware, err := middlewares.NewRateLimitMiddleware(limits)
	if err != nil {
		t.Fatalf("Error creating middleware: %v", err)
	}

	handler := rateLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()

	// First and second request should pass
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", rr.Code)
	}

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", rr.Code)
	}

	// Third request should fail due to 1-second window limit
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("expected status TooManyRequests, got %v", rr.Code)
	}
}
