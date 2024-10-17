package middlewares

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

const DEFAULT_CACHE_TTL = time.Minute * 1
const CLEANUP_CACHE_INTERVAL = time.Minute * 10

var responseCache = cache.New(DEFAULT_CACHE_TTL, CLEANUP_CACHE_INTERVAL) // Default cache with a placeholder TTL

type CachedResponse struct {
	statusCode int
	body       []byte
	headers    map[string]string
}

func NewCacheMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if response  response is already cached
			cachedResponse, found := responseCache.Get(r.URL.String())
			if found {
				response := cachedResponse.(CachedResponse)
				// TODO: cache response headers
				for header, value := range response.headers {
					w.Header().Set(header, value)
				}
				w.WriteHeader(response.statusCode)
				w.Write(response.body)
				fmt.Println("Returning cached")
				return
			}

			// Serve the next handler and capture the response
			rc := NewResponseCapture(w)
			next.ServeHTTP(rc, r)

			// Get cache control headers
			cacheControl := rc.Header().Get("Cache-Control")
			maxAge := getCacheMaxAge(cacheControl)
			expires := getCacheExpires(rc.Header().Get("Expires"))

			// Determine TTL based on cache headers
			ttl := time.Second * 0
			if maxAge > 0 {
				ttl = time.Duration(maxAge) * time.Second
			} else if !expires.IsZero() {
				ttl = time.Until(expires)
			}

			// Cache the response if it's cacheable
			if ttl > 0 {
				cachedResponse := CachedResponse{statusCode: rc.status, body: rc.buffer.Bytes()}
				responseCache.Set(r.URL.String(), cachedResponse, ttl)
			}

			// Write the captured response (original or cached)
			rc.Flush()
		})
	}
}

func getCacheMaxAge(cacheControl string) int {
	for _, directive := range strings.Split(cacheControl, ",") {
		directive = strings.TrimSpace(directive)
		if strings.HasPrefix(directive, "max-age=") {
			maxAge, err := strconv.Atoi(strings.TrimPrefix(directive, "max-age="))
			if err == nil {
				return maxAge
			}
		}
	}
	return 0
}

func getCacheExpires(expiresHeader string) time.Time {
	expires, err := time.Parse(time.RFC1123, expiresHeader)
	if err != nil {
		return time.Time{}
	}
	return expires
}
