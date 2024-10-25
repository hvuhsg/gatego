package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/hvuhsg/gatego/internal/config"
)

func TestNewBalancer(t *testing.T) {
	service := config.Service{}
	path := config.Path{
		Backend: &config.Backend{
			BalancePolicy: "round-robin",
			Servers: []struct {
				URL    string "yaml:\"url\""
				Weight uint   "yaml:\"weight\""
			}{
				{URL: "http://localhost:8001", Weight: 1},
				{URL: "http://localhost:8002", Weight: 2},
			},
		},
	}

	balancer, err := NewBalancer(service, path)
	if err != nil {
		t.Fatalf("Failed to create balancer: %v", err)
	}

	if balancer == nil {
		t.Fatal("Balancer is nil")
	}
}

func TestRoundRobinPolicy(t *testing.T) {
	servers := []ServerAndWeight{
		{server: createDummyProxy("http://localhost:8001/"), weight: 1, url: "http://localhost:8001/"},
		{server: createDummyProxy("http://localhost:8002/"), weight: 1, url: "http://localhost:8002/"},
	}

	policy := NewRoundRobinPolicy(servers)

	// Test the round-robin behavior
	expectedOrder := []string{"http://localhost:8001/", "http://localhost:8002/", "http://localhost:8001/", "http://localhost:8002/"}
	for i, expected := range expectedOrder {
		server := policy.GetNext()
		if server.Director == nil {
			t.Fatalf("Server %d is nil", i)
		}
		serverURL := getProxyURL(server)
		if serverURL != expected {
			t.Errorf("index = %d Expected server %s, got %s", i, expected, serverURL)
		}
	}
}

func TestRandomPolicy(t *testing.T) {
	servers := []ServerAndWeight{
		{server: createDummyProxy("http://localhost:8001"), weight: 1, url: "http://localhost:8001"},
		{server: createDummyProxy("http://localhost:8002"), weight: 1, url: "http://localhost:8002"},
	}

	policy := NewRandomPolicy(servers)

	// Test that we get a valid server (we can't test randomness easily)
	for i := 0; i < 10; i++ {
		server := policy.GetNext()
		if server == nil {
			t.Fatal("Got nil server from RandomPolicy")
		}
	}
}

func TestLeastLatencyPolicy(t *testing.T) {
	// Create mock servers
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Slow response from server 1"))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Fast response from server 2"))
	}))
	defer server2.Close()

	servers := []ServerAndWeight{
		{server: httputil.NewSingleHostReverseProxy(mustParseURL(server1.URL)), weight: 1, url: server1.URL},
		{server: httputil.NewSingleHostReverseProxy(mustParseURL(server2.URL)), weight: 1, url: server2.URL},
	}

	policy := NewLeastLatencyPolicy(servers)

	// Initially, all servers should have 0 latency
	server := policy.GetNext()
	if server == nil {
		t.Fatal("Got nil server from LeastLatencyPolicy")
	}

	// Simulate a request and update latency
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", server1.URL, nil)
	server.ServeHTTP(w, r)

	// The policy should now prefer the fast second server
	server = policy.GetNext()
	serverURL := strings.TrimSuffix(getProxyURL(server), "/")
	if serverURL != strings.TrimSuffix(server2.URL, "/") {
		t.Errorf("LeastLatencyPolicy did not choose the server with least latency Got %s Want %s", serverURL, server2.URL)
	}
}

func TestBalancerServeHTTP(t *testing.T) {
	// Create mock servers
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Response from server 1"))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Response from server 2"))
	}))
	defer server2.Close()

	// Create ServerAndWeight structs using the mock servers
	servers := []ServerAndWeight{
		{server: httputil.NewSingleHostReverseProxy(mustParseURL(server1.URL)), weight: 1, url: server1.URL},
		{server: httputil.NewSingleHostReverseProxy(mustParseURL(server2.URL)), weight: 1, url: server2.URL},
	}

	policy := NewRoundRobinPolicy(servers)
	balancer := &Balancer{policy: policy}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://example.com", nil)

	balancer.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "http://example.com", nil)

	balancer.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

// Helper function to create a dummy reverse proxy
func createDummyProxy(targetURL string) *httputil.ReverseProxy {
	url, _ := url.Parse(targetURL)
	return httputil.NewSingleHostReverseProxy(url)
}

// Helper function to get the target URL of a reverse proxy
func getProxyURL(proxy *httputil.ReverseProxy) string {
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	proxy.Director(req)
	return req.URL.String()
}

// Helper function to parse URL and panic on error
func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}
