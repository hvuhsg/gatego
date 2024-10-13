package handlers

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/hvuhsg/gatego/config"
)

type Proxy struct {
	proxy *httputil.ReverseProxy
}

func NewProxy(service config.Service, path config.Path) (Proxy, error) {
	serviceURL, err := url.Parse(*path.Destination)
	if err != nil {
		return Proxy{}, err
	}

	proxy := httputil.NewSingleHostReverseProxy(serviceURL)

	server := Proxy{proxy: proxy}
	return server, nil
}

func (p Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.proxy.ServeHTTP(w, r)
}
