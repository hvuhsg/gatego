package middlewares

import (
	"net/http"
	"strconv"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/json"
	"github.com/tdewolff/minify/v2/svg"
	"github.com/tdewolff/minify/v2/xml"
)

type MinifyConfig struct {
	ALL  bool `yaml:"all"`
	JS   bool `yaml:"js"`
	CSS  bool `yaml:"css"`
	HTML bool `yaml:"html"`
	JSON bool `yaml:"json"`
	SVG  bool `yaml:"svg"`
	XML  bool `yaml:"xml"`
}

func NewMinifyMiddleware(config MinifyConfig) Middleware {
	m := minify.New()

	// Add minifiers for the different content types
	if config.HTML || config.ALL {
		m.AddFunc("text/html", html.Minify)
	}
	if config.CSS || config.ALL {
		m.AddFunc("text/css", css.Minify)
	}
	if config.JS || config.ALL {
		m.AddFunc("application/javascript", js.Minify)
	}
	if config.JSON || config.ALL {
		m.AddFunc("application/json", json.Minify)
	}
	if config.SVG || config.ALL {
		m.AddFunc("image/svg+xml", svg.Minify)
	}
	if config.XML || config.ALL {
		m.AddFunc("application/xml", xml.Minify)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a custom ResponseWriter to capture the response
			rw := NewResponseCapture(w)

			// Serve the next handler
			next.ServeHTTP(rw, r)

			// Get the content type of the response
			contentType := rw.Header().Get("Content-Type")

			minifiedContent, err := m.Bytes(contentType, rw.buffer.Bytes())
			if err != nil {
				rw.Flush() // Return the original response
				return
			}

			// Write the minified content to the response
			w.Header().Set("Content-Length", strconv.Itoa(len(minifiedContent)))

			if rw.status > 0 {
				w.WriteHeader(rw.status)
			}

			w.Write(minifiedContent)
		})
	}
}
