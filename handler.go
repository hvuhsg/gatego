package gatego

import (
	"errors"
	"net/http"
	"os"
	"slices"

	"github.com/hvuhsg/gatego/config"
	"github.com/hvuhsg/gatego/handlers"
	"github.com/hvuhsg/gatego/middlewares"
)

var ErrUnsupportedBaseHandler = errors.New("base handler unsupported")

// func loggingMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		rc := middlewares.NewResponseCapture(w)
// 		next.ServeHTTP(rc, r)

// 		w.WriteHeader(rc.Status())
// 		w.Write(rc.Buffer())
// 	})
// }

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

func NewHandler(service config.Service, path config.Path) (http.Handler, error) {
	handler, err := GetBaseHandler(service, path)
	if err != nil {
		return nil, err
	}

	handlerWithMiddlewares := middlewares.NewHandlerWithMiddleware(handler)

	handlerWithMiddlewares.Add(middlewares.NewLoggingMiddleware(os.Stdout))

	if path.Timeout == 0 {
		path.Timeout = config.DefaultTimeout
	}
	handlerWithMiddlewares.Add(middlewares.NewTimeoutMiddleware(path.Timeout))

	if path.MaxSize == 0 {
		path.MaxSize = config.DefaultMaxRequestSize
	}
	handlerWithMiddlewares.Add(middlewares.NewRequestSizeLimitMiddleware(path.MaxSize))

	if len(path.RateLimits) > 0 {
		ratelimiter, err := middlewares.NewRateLimitMiddleware(path.RateLimits)
		if err != nil {
			return nil, err
		}
		handlerWithMiddlewares.Add(ratelimiter)
	}

	if path.Headers != nil {
		handlerWithMiddlewares.Add(middlewares.NewAddHeadersMiddleware(*path.Headers))
	}

	if path.Gzip != nil && *path.Gzip {
		handlerWithMiddlewares.Add(middlewares.GzipMiddleware)
	}

	if len(path.OmitHeaders) > 0 {
		handlerWithMiddlewares.Add(middlewares.NewOmitHeadersMiddleware(path.OmitHeaders))
	}

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

	if path.OpenAPI != nil {
		openapiMiddleware, err := middlewares.NewOpenAPIValidationMiddleware(*path.OpenAPI)
		if err != nil {
			return nil, err
		}
		handlerWithMiddlewares.Add(openapiMiddleware)
	}

	if path.Cache {
		handlerWithMiddlewares.Add(middlewares.NewCacheMiddleware())
	}

	// handlerWithMiddlewares.Add(loggingMiddleware)

	return handlerWithMiddlewares, nil
}
