package tracker

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type Tracker interface {
	GetTrackerID(*http.Request) string
	SetTracker(http.ResponseWriter) (string, error)
	RemoveTracker(*http.Request)
}

type cookieTracker struct {
	cookieName    string
	trackerMaxAge int
	secureCookie  bool
}

func NewCookieTracker(cookieName string, maxAge int, isSecure bool) cookieTracker {
	return cookieTracker{cookieName: cookieName, trackerMaxAge: maxAge, secureCookie: isSecure}
}

// Get the tracker id from request or return empty string if not found
func (ct cookieTracker) GetTrackerID(r *http.Request) string {
	cookie, err := r.Cookie(ct.cookieName)

	if err != nil {
		return ""
	}

	return cookie.Value
}

// Set tracer into response and return the tracker id
func (ct cookieTracker) SetTracker(w http.ResponseWriter) (string, error) {
	traceID, err := generateTraceID()
	if err != nil {
		return "", err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     ct.cookieName,
		Value:    traceID,
		Path:     "/",
		MaxAge:   ct.trackerMaxAge,
		HttpOnly: true,
		Secure:   ct.secureCookie,
		SameSite: http.SameSiteLaxMode,
	})

	return traceID, nil
}

func (ct cookieTracker) RemoveTracker(r *http.Request) {
	// Get existing cookies
	oldCookies := r.Cookies()

	// Create new headers without the cookie we want to remove
	r.Header.Del("Cookie")

	// Add back all cookies except the one we want to remove
	for _, cookie := range oldCookies {
		if cookie.Name != ct.cookieName {
			r.AddCookie(cookie)
		}
	}
}

func generateTraceID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
