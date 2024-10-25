package middlewares

import (
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"go.opentelemetry.io/otel/trace"
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
			span := trace.SpanFromContext(r.Context())

			route, pathParams, err := router.FindRoute(r)
			if err != nil {
				span.AddEvent("Request path not found in openapi spec")
				http.Error(w, fmt.Sprintf("Error finding route: %v", err), http.StatusBadRequest)
				return
			}

			requestValidationInput := &openapi3filter.RequestValidationInput{
				Request:    r,
				PathParams: pathParams,
				Route:      route,
			}

			if err := openapi3filter.ValidateRequest(r.Context(), requestValidationInput); err != nil {
				span.AddEvent(fmt.Sprintf("Error while validating request with openapi spec. err = %v", err))
				http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
				return
			}

			span.AddEvent("Request validated by openapi spec")

			rc := NewRecorder()
			next.ServeHTTP(rc, r)

			responseValidationInput := &openapi3filter.ResponseValidationInput{
				RequestValidationInput: requestValidationInput,
				Status:                 rc.Result().StatusCode,
				Header:                 rc.Header(),
			}

			if rc.Body.Bytes() != nil {
				responseValidationInput.SetBodyBytes(rc.Body.Bytes())
			}

			if err := openapi3filter.ValidateResponse(r.Context(), responseValidationInput); err != nil {
				span.AddEvent(fmt.Sprintf("Error while validating response with openapi spec. err = %v", err))
				http.Error(w, fmt.Sprintf("Invalid response: %v", err), http.StatusInternalServerError)
				return
			}

			span.AddEvent("Response validated by openapi spec")

			rc.WriteHeadersTo(w)
			w.WriteHeader(rc.Result().StatusCode)
			if rc.Body.Bytes() != nil {
				w.Write(rc.Body.Bytes())
			}
		})
	}, nil
}
