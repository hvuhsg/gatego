package gatego

import (
	"context"
	"errors"
	"net/http"
	"os"
	"slices"

	"github.com/hvuhsg/gatego/config"
	"github.com/hvuhsg/gatego/handlers"
	"github.com/hvuhsg/gatego/middlewares"
)

var ErrUnsupportedBaseHandler = errors.New("base handler unsupported")

func GetBaseHandler(service config.Service, path config.Path) (http.Handler, error) {
	if path.Destination != nil && *path.Destination != "" {
		return handlers.NewProxy(service, path)
	} else if path.Directory != nil && *path.Directory != "" {
		handler := handlers.NewFiles(*path.Directory, path.Path)
		return handler, nil
	} else if path.Backend != nil {
		return handlers.NewBalancer(service, path)
	} else {
		// Should not be reached (early validation should prevent it)
		return nil, ErrUnsupportedBaseHandler
	}
}

func NewHandler(ctx context.Context, useOtel bool, service config.Service, path config.Path) (http.Handler, error) {
	handler, err := GetBaseHandler(service, path)
	if err != nil {
		return nil, err
	}

	handlerWithMiddlewares := middlewares.NewHandlerWithMiddleware(handler)

	handlerWithMiddlewares.Add(middlewares.NewLoggingMiddleware(os.Stdout))

	// Open Telemetry
	if useOtel {
		otelMiddleware, err := middlewares.NewOpenTelemetryMiddleware(
			ctx,
			middlewares.OTELConfig{
				ServiceDomain: service.Domain,
				BasePath:      path.Path,
			},
		)
		if err != nil {
			return nil, err
		}
		handlerWithMiddlewares.Add(otelMiddleware)
	}

	// Timeout
	if path.Timeout == 0 {
		path.Timeout = config.DefaultTimeout
	}
	handlerWithMiddlewares.Add(middlewares.NewTimeoutMiddleware(path.Timeout))

	// Max request size
	if path.MaxSize == 0 {
		path.MaxSize = config.DefaultMaxRequestSize
	}
	handlerWithMiddlewares.Add(middlewares.NewRequestSizeLimitMiddleware(path.MaxSize))

	// Rate limits
	if len(path.RateLimits) > 0 {
		ratelimiter, err := middlewares.NewRateLimitMiddleware(path.RateLimits)
		if err != nil {
			return nil, err
		}
		handlerWithMiddlewares.Add(ratelimiter)
	}

	// Add headers
	if path.Headers != nil {
		handlerWithMiddlewares.Add(middlewares.NewAddHeadersMiddleware(*path.Headers))
	}

	// GZIP compression
	if path.Gzip != nil && *path.Gzip {
		handlerWithMiddlewares.Add(middlewares.GzipMiddleware)
	}

	// Remove response headers
	if len(path.OmitHeaders) > 0 {
		handlerWithMiddlewares.Add(middlewares.NewOmitHeadersMiddleware(path.OmitHeaders))
	}

	// Minify files
	minifyConfig := middlewares.MinifyConfig{
		ALL:  slices.Contains(path.Minify, "all"),
		JS:   slices.Contains(path.Minify, "js"),
		HTML: slices.Contains(path.Minify, "html"),
		CSS:  slices.Contains(path.Minify, "css"),
		JSON: slices.Contains(path.Minify, "json"),
		SVG:  slices.Contains(path.Minify, "svg"),
		XML:  slices.Contains(path.Minify, "xml"),
	}
	handlerWithMiddlewares.Add(middlewares.NewMinifyMiddleware(minifyConfig))

	// OpenAPI validation
	if path.OpenAPI != nil {
		openapiMiddleware, err := middlewares.NewOpenAPIValidationMiddleware(*path.OpenAPI)
		if err != nil {
			return nil, err
		}
		handlerWithMiddlewares.Add(openapiMiddleware)
	}

	// Response cache
	if path.Cache {
		handlerWithMiddlewares.Add(middlewares.NewCacheMiddleware())
	}

	return handlerWithMiddlewares, nil
}
