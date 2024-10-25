package middlewares

import "net/http"

type Middleware func(http.Handler) http.Handler

type HandlerWithMiddleware struct {
	finalHandler http.Handler
	middlewares  []Middleware
}

func NewHandlerWithMiddleware(handler http.Handler) *HandlerWithMiddleware {
	return &HandlerWithMiddleware{
		finalHandler: handler,
		middlewares:  []Middleware{},
	}
}

func (h *HandlerWithMiddleware) Add(middleware Middleware) {
	h.middlewares = append(h.middlewares, middleware)
}

func (h *HandlerWithMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Chain the middlewares around the final handler
	handler := h.finalHandler
	for i := len(h.middlewares) - 1; i >= 0; i-- {
		handler = h.middlewares[i](handler)
	}
	handler.ServeHTTP(w, r)
}
