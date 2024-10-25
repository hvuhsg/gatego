package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hvuhsg/gatego/internal/middlewares"
)

// Helper function to create a basic next handler that returns content with a specific content type
func createHandler(contentType, content string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	})
}

// TestMinifyMiddleware_HTML tests HTML minification
func TestMinifyMiddleware_HTML(t *testing.T) {
	handler := createHandler("text/html", "<html>  <body>    <h1>Hello World</h1>  </body></html>")

	config := middlewares.MinifyConfig{HTML: true}
	middleware := middlewares.NewMinifyMiddleware(config)
	minifiedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	minifiedHandler.ServeHTTP(rr, req)

	expected := "<h1>Hello World</h1>"
	if rr.Body.String() != expected {
		t.Errorf("expected '%s', got '%s'", expected, rr.Body.String())
	}
}

// TestMinifyMiddleware_CSS tests CSS minification
func TestMinifyMiddleware_CSS(t *testing.T) {
	handler := createHandler("text/css", "body { color: red; }")

	config := middlewares.MinifyConfig{CSS: true}
	middleware := middlewares.NewMinifyMiddleware(config)
	minifiedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	minifiedHandler.ServeHTTP(rr, req)

	expected := "body{color:red}"
	if rr.Body.String() != expected {
		t.Errorf("expected '%s', got '%s'", expected, rr.Body.String())
	}
}

// TestMinifyMiddleware_JS tests JS minification
func TestMinifyMiddleware_JS(t *testing.T) {
	handler := createHandler("application/javascript", "function test() { return 1; }")

	config := middlewares.MinifyConfig{JS: true}
	middleware := middlewares.NewMinifyMiddleware(config)
	minifiedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	minifiedHandler.ServeHTTP(rr, req)

	expected := "function test(){return 1}"
	if rr.Body.String() != expected {
		t.Errorf("expected '%s', got '%s'", expected, rr.Body.String())
	}
}

// TestMinifyMiddleware_JSON tests JSON minification
func TestMinifyMiddleware_JSON(t *testing.T) {
	handler := createHandler("application/json", `{
		"name": "John",
		"age": 30
	}`)

	config := middlewares.MinifyConfig{JSON: true}
	middleware := middlewares.NewMinifyMiddleware(config)
	minifiedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	minifiedHandler.ServeHTTP(rr, req)

	expected := `{"name":"John","age":30}`
	if rr.Body.String() != expected {
		t.Errorf("expected '%s', got '%s'", expected, rr.Body.String())
	}
}

// TestMinifyMiddleware_SkipUnsupported tests that unsupported content types are not minified
func TestMinifyMiddleware_SkipUnsupported(t *testing.T) {
	handler := createHandler("text/plain", "This is a plain text file.")

	config := middlewares.MinifyConfig{HTML: true, CSS: true, JS: true}
	middleware := middlewares.NewMinifyMiddleware(config)
	minifiedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	minifiedHandler.ServeHTTP(rr, req)

	expected := "This is a plain text file."
	if rr.Body.String() != expected {
		t.Errorf("expected '%s', got '%s'", expected, rr.Body.String())
	}
}
