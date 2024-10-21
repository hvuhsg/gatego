package handlers

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/hvuhsg/gatego/config"
	"github.com/hvuhsg/gatego/contextvalues"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
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
	tracer := contextvalues.TracerFromContext(r.Context())
	if tracer != nil {
		ctx, span := tracer.Start(r.Context(), "request.upstream")
		span.SetAttributes(semconv.HTTPServerAttributesFromHTTPRequest(r.Host, r.URL.Path, r)...)
		r = r.WithContext(ctx)
		defer span.End()
	}
	p.proxy.ServeHTTP(w, r)
}
