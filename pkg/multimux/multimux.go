// This package implement a mutil-mux an http handler
// that acts as seprate http.ServeMux for each registred host

package multimux

import (
	"net/http"
	"strings"
)

type MultiMux struct {
	Hosts map[string]*http.ServeMux
}

func NewMultiMux() *MultiMux {
	hosts := make(map[string]*http.ServeMux)
	return &MultiMux{Hosts: hosts}
}

func (mm *MultiMux) RegisterHandler(host string, pattern string, handler http.Handler) {
	cleanedHost := cleanHost(host)
	mux, exists := mm.Hosts[cleanedHost]

	if !exists {
		mux = http.NewServeMux()
		mm.Hosts[cleanedHost] = mux
	}

	cleanedPattern := strings.ToLower(pattern)

	mux.Handle(cleanedPattern, handler)
}

func (mm *MultiMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	cleanedHost := cleanHost(host)
	mux, exists := mm.Hosts[cleanedHost]

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	mux.ServeHTTP(w, r)
}

func cleanHost(domain string) string {
	return removePort(strings.ToLower(domain))
}

func removePort(addr string) string {
	if i := strings.LastIndex(addr, ":"); i != -1 {
		return addr[:i]
	}
	return addr
}
