package tracker

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewCookieTracker(t *testing.T) {
	tracker := NewCookieTracker("testTracker", 3600, true)

	if tracker.cookieName != "testTracker" {
		t.Errorf("Expected cookieName to be 'testTracker', got %s", tracker.cookieName)
	}
	if tracker.trackerMaxAge != 3600 {
		t.Errorf("Expected trackerMaxAge to be 3600, got %d", tracker.trackerMaxAge)
	}
	if !tracker.secureCookie {
		t.Errorf("Expected secureCookie to be true")
	}
}

func TestGenerateTraceID(t *testing.T) {
	traceID1, err1 := generateTraceID()
	if err1 != nil {
		t.Fatalf("Unexpected error generating trace ID: %v", err1)
	}

	traceID2, err2 := generateTraceID()
	if err2 != nil {
		t.Fatalf("Unexpected error generating trace ID: %v", err2)
	}

	if len(traceID1) != 32 {
		t.Errorf("Expected trace ID length to be 32, got %d", len(traceID1))
	}

	if traceID1 == traceID2 {
		t.Error("Generated trace IDs should be unique")
	}
}

func TestSetTracker(t *testing.T) {
	tracker := NewCookieTracker("testTracker", 3600, true)

	// Create a test response writer
	w := httptest.NewRecorder()

	// Set tracker
	traceID, err := tracker.SetTracker(w)
	if err != nil {
		t.Fatalf("Unexpected error setting tracker: %v", err)
	}

	// Check response headers
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]
	if cookie.Name != "testTracker" {
		t.Errorf("Expected cookie name 'testTracker', got %s", cookie.Name)
	}
	if cookie.Value != traceID {
		t.Errorf("Cookie value does not match returned trace ID")
	}
	if cookie.Path != "/" {
		t.Errorf("Expected cookie path '/', got %s", cookie.Path)
	}
	if cookie.MaxAge != 3600 {
		t.Errorf("Expected MaxAge 3600, got %d", cookie.MaxAge)
	}
	if !cookie.HttpOnly {
		t.Errorf("Expected HttpOnly to be true")
	}
}

func TestGetTrackerID(t *testing.T) {
	tracker := NewCookieTracker("testTracker", 3600, true)

	// Test request without cookie
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	trackerID1 := tracker.GetTrackerID(req1)
	if trackerID1 != "" {
		t.Errorf("Expected empty string when no cookie exists, got %s", trackerID1)
	}

	// Test request with cookie
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.AddCookie(&http.Cookie{
		Name:  "testTracker",
		Value: "test-trace-id",
	})

	trackerID2 := tracker.GetTrackerID(req2)
	if trackerID2 != "test-trace-id" {
		t.Errorf("Expected 'test-trace-id', got %s", trackerID2)
	}
}

func TestRemoveTracker(t *testing.T) {
	tracker := NewCookieTracker("testTracker", 3600, true)

	// Create a request with multiple cookies
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "testTracker", Value: "remove-me"})
	req.AddCookie(&http.Cookie{Name: "otherCookie", Value: "keep-me"})

	// Remove the specific tracker cookie
	tracker.RemoveTracker(req)

	// Check that the cookie header has been modified
	cookieHeader := req.Header.Get("Cookie")
	if strings.Contains(cookieHeader, "testTracker") {
		t.Errorf("testTracker cookie should have been removed")
	}
	if !strings.Contains(cookieHeader, "otherCookie=keep-me") {
		t.Errorf("Other cookies should be preserved")
	}
}

// Benchmark trace ID generation
func BenchmarkGenerateTraceID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateTraceID()
	}
}
