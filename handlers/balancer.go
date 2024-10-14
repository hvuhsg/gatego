package handlers

import (
	"math"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/hvuhsg/gatego/config"
)

type ServerAndWeight struct {
	server *httputil.ReverseProxy
	weight int
	url    string
}

type BalancePolicy interface {
	GetNext() *httputil.ReverseProxy
}

type Balancer struct {
	policy BalancePolicy
}

func NewBalancer(service config.Service, path config.Path) (*Balancer, error) {
	serversConfig := path.Backend.Servers

	serversAndWeights := make([]ServerAndWeight, 0, len(serversConfig))
	for _, serverConfig := range serversConfig {
		serverURL, err := url.Parse(serverConfig.URL)
		if err != nil {
			return &Balancer{}, err
		}

		server := httputil.NewSingleHostReverseProxy(serverURL)

		serverWeight := int(serverConfig.Weight)
		if serverWeight < 1 {
			serverWeight = 1
		}
		serversAndWeights = append(serversAndWeights, ServerAndWeight{server: server, weight: serverWeight, url: serverConfig.URL})
	}

	var policy BalancePolicy
	switch path.Backend.BalancePolicy {
	case "round-robin":
		policy = NewRoundRobinPolicy(serversAndWeights)
	case "random":
		policy = NewRandomPolicy(serversAndWeights)
	case "least-latency":
		policy = NewLeastLatencyPolicy(serversAndWeights)
	}

	balancer := Balancer{policy: policy}

	return &balancer, nil
}

func (b *Balancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	proxy := b.policy.GetNext()
	proxy.ServeHTTP(w, r)
}

type RoundRobinPolicy struct {
	current    int
	weightsSum int
	servers    []ServerAndWeight
}

func NewRoundRobinPolicy(servers []ServerAndWeight) *RoundRobinPolicy {
	weightsSum := 0
	for _, server := range servers {
		weightsSum += server.weight
	}

	policy := &RoundRobinPolicy{current: 0, weightsSum: weightsSum, servers: servers}
	return policy
}

// The servers provided must be provided in the same order for accurate results
func (rrp *RoundRobinPolicy) GetNext() *httputil.ReverseProxy {
	serverIndex := rrp.current

	for _, server := range rrp.servers {
		serverIndex -= server.weight
		if serverIndex < 0 {
			rrp.current += 1
			return server.server
		}
	}

	rrp.current = (rrp.current % rrp.weightsSum) + 1
	return rrp.servers[0].server
}

type RandomPolicy struct {
	weightsSum int
	servers    []ServerAndWeight
}

func NewRandomPolicy(servers []ServerAndWeight) *RandomPolicy {
	weightsSum := 0
	for _, server := range servers {
		weightsSum += server.weight
	}

	return &RandomPolicy{weightsSum: weightsSum, servers: servers}
}

func (rp *RandomPolicy) GetNext() *httputil.ReverseProxy {
	randomServerIndex := rand.Intn(rp.weightsSum)

	for _, server := range rp.servers {
		randomServerIndex -= server.weight
		if randomServerIndex <= 0 {
			return server.server
		}
	}

	return rp.servers[0].server
}

type LeastLatencyPolicy struct {
	serversLatency map[string]int64
	servers        []ServerAndWeight
}

func NewLeastLatencyPolicy(serversAndURLs []ServerAndWeight) *LeastLatencyPolicy {
	serversLatency := make(map[string]int64, len(serversAndURLs))

	for _, serverAndWeight := range serversAndURLs {
		serversLatency[serverAndWeight.url] = 0
	}

	return &LeastLatencyPolicy{servers: serversAndURLs, serversLatency: serversLatency}
}

func (llp *LeastLatencyPolicy) GetNext() *httputil.ReverseProxy {

	bestServerURL := llp.servers[0].url
	var bestLatency int64 = math.MaxInt64

	for url, latency := range llp.serversLatency {
		if latency < bestLatency {
			bestServerURL = url
			bestLatency = latency
		}
	}

	var chosenServer ServerAndWeight
	for _, server := range llp.servers {
		if server.url == bestServerURL {
			chosenServer = server
			break
		}
	}

	// TODO: use decaing latency for extream latency conditions

	startTime := time.Now().UnixMicro()
	chosenServer.server.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		llp.serversLatency[chosenServer.url] = time.Now().UnixMicro() - startTime
	}
	chosenServer.server.ModifyResponse = func(r *http.Response) error {
		llp.serversLatency[chosenServer.url] = time.Now().UnixMicro() - startTime
		return nil
	}

	return chosenServer.server
}
