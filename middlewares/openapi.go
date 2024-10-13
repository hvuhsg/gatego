package middlewares

import (
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func NewOpenAPIValidationMiddleware(specPath string) (Middleware, error) {
	loader := &openapi3.Loader{IsExternalRefsAllowed: true}
	doc, err := loader.LoadFromFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("error loading OpenAPI spec: %w", err)
	}

	if err := doc.Validate(loader.Context); err != nil {
		return nil, fmt.Errorf("error validating OpenAPI spec: %w", err)
	}

	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		return nil, fmt.Errorf("error creating router: %w", err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route, pathParams, err := router.FindRoute(r)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error finding route: %v", err), http.StatusBadRequest)
				return
			}

			requestValidationInput := &openapi3filter.RequestValidationInput{
				Request:    r,
				PathParams: pathParams,
				Route:      route,
			}

			if err := openapi3filter.ValidateRequest(r.Context(), requestValidationInput); err != nil {
				fmt.Println("invalid request", err)
				http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
				return
			}

			rc := NewResponseCapture(w)
			next.ServeHTTP(rc, r)

			responseValidationInput := &openapi3filter.ResponseValidationInput{
				RequestValidationInput: requestValidationInput,
				Status:                 rc.status,
				Header:                 rc.Header(),
			}

			if rc.buffer.Bytes() != nil {
				responseValidationInput.SetBodyBytes(rc.buffer.Bytes())
			}

			if err := openapi3filter.ValidateResponse(r.Context(), responseValidationInput); err != nil {
				http.Error(w, fmt.Sprintf("Invalid response: %v", err), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(rc.status)
			if rc.buffer.Bytes() != nil {
				w.Write(rc.buffer.Bytes())
			}
		})
	}, nil
}
