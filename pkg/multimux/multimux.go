// This package implement a mutil-mux an http handler
// that acts as seprate http.ServeMux for each registred host

package multimux

import (
	"net/http"
	"strings"
	"sync"
)

type MultiMux struct {
	Hosts sync.Map
}

func NewMultiMux() *MultiMux {
	return &MultiMux{Hosts: sync.Map{}}
}

func (mm *MultiMux) RegisterHandler(host string, pattern string, handler http.Handler) {
	cleanedHost := cleanHost(host)
	muxAny, _ := mm.Hosts.LoadOrStore(cleanedHost, http.NewServeMux())
	mux := muxAny.(*http.ServeMux)

	cleanedPattern := strings.ToLower(pattern)

	mux.Handle(cleanedPattern, handler)
}

func (mm *MultiMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	cleanedHost := cleanHost(host)
	muxAny, exists := mm.Hosts.Load(cleanedHost)

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	mux := muxAny.(*http.ServeMux)
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
