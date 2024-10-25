package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

var SupportedZones = []string{"ip"}
var ErrZoneNotSupported = errors.New("rate limit zone is not supported")

type RateLimiter struct {
	limiters sync.Map
}

type LimitConfig struct {
	Zone     string
	Requests int
	Per      time.Duration
}

func (lc LimitConfig) GetKey(r *http.Request) (key string, err error) {
	err = nil
	switch lc.Zone {
	case "ip":
		parts := strings.Split(r.RemoteAddr, ":")
		ip := parts[0]
		key = "ip:" + ip
	default:
		err = errors.New("rate limit zone is not supported")
	}
	key = strconv.Itoa(int(lc.Per.Seconds())) + "|" + strconv.Itoa(lc.Requests) + "!" + key
	return
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{}
}

func (rl *RateLimiter) addLimiter(key string, limit rate.Limit, burst int) {
	limiter := rate.NewLimiter(limit, burst)
	rl.limiters.Store(key, limiter)
}

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	if limiter, ok := rl.limiters.Load(key); ok {
		return limiter.(*rate.Limiter)
	}
	return nil
}

func ParseLimitConfig(config string) (LimitConfig, error) {
	parts := strings.Split(config, "-")
	if len(parts) != 2 {
		return LimitConfig{}, fmt.Errorf("invalid limit config: %s", config)
	}
	zone := parts[0]
	if !slices.Contains(SupportedZones, strings.ToLower(zone)) {
		return LimitConfig{}, ErrZoneNotSupported
	}

	limitParts := strings.Split(parts[1], "/")
	if len(limitParts) != 2 {
		return LimitConfig{}, fmt.Errorf("invalid limit config: %s", config)
	}

	requests, err := strconv.Atoi(limitParts[0])
	if err != nil {
		return LimitConfig{}, fmt.Errorf("invalid requests number: %s", limitParts[0])
	}

	var duration time.Duration
	switch limitParts[1] {
	case "s":
		duration = time.Second
	case "m":
		duration = time.Minute
	case "h":
		duration = time.Hour
	case "d":
		duration = time.Hour * 24
	default:
		return LimitConfig{}, fmt.Errorf("invalid time unit: %s", limitParts[1])
	}

	return LimitConfig{
		Zone:     zone,
		Requests: requests,
		Per:      duration,
	}, nil
}

func NewRateLimitMiddleware(limits []string) (func(http.Handler) http.Handler, error) {
	rateLimiter := NewRateLimiter()

	// Pre-process ratelimit configs
	parsedLimits := make([]LimitConfig, 0, len(limits))
	for _, limit := range limits {
		parsedLimit, err := ParseLimitConfig(limit)
		if err != nil {
			return nil, err
		}
		parsedLimits = append(parsedLimits, parsedLimit)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, config := range parsedLimits {
				key, err := config.GetKey(r)
				if err != nil {
					// Should never reach here (validation should prevent it)
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}

				limiter := rateLimiter.getLimiter(key)
				if limiter == nil {
					rateLimiter.addLimiter(key, rate.Every(config.Per), config.Requests)
					limiter = rateLimiter.getLimiter(key)
				}

				if !limiter.Allow() {
					setRateLimitHeaders(w, limiter, config)
					http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}, nil
}

func setRateLimitHeaders(w http.ResponseWriter, limiter *rate.Limiter, config LimitConfig) {
	now := time.Now()
	limit := config.Requests
	remaining := int(limiter.Tokens())
	reset := now.Add(config.Per).Unix()

	w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
	w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
	w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", reset))
}
