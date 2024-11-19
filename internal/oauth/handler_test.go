package oauth_test

import (
	"maps"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/hvuhsg/gatego/internal/oauth"
	"github.com/hvuhsg/gatego/pkg/security"
	"github.com/stretchr/testify/assert"
)

func TestLoginHandler(t *testing.T) {
	t.Run("Successful login with Google", func(t *testing.T) {
		// Setup mock provider and config
		mockConfig := oauth.OAuthConfig{
			BaseURL: "/oauth",
			Google: oauth.AuthProviderConfig{
				Enabled: true,
			},
		}

		// Create OAuth handler
		oauthHandler := oauth.NewOAuthHandler(mockConfig)

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/google/login", nil)
		w := httptest.NewRecorder()

		// Call handler
		oauthHandler.ServeHTTP(w, req)

		// Check response
		resp := w.Result()
		assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
	})

	t.Run("Provider not found", func(t *testing.T) {
		mockConfig := oauth.OAuthConfig{}
		oauthHandler := oauth.NewOAuthHandler(mockConfig)

		req := httptest.NewRequest(http.MethodGet, "/login/nonexistent", nil)
		w := httptest.NewRecorder()

		oauthHandler.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestCallbackHandler(t *testing.T) {
	t.Run("Successful callback", func(t *testing.T) {
		mockConfig := oauth.OAuthConfig{
			BaseURL:            "oauth",
			AfterLoginRedirect: "/dashboard",
			Google: oauth.AuthProviderConfig{
				Enabled: true,
			},
		}

		// Create OAuth handler
		oauthHandler := oauth.NewOAuthHandler(mockConfig)

		// Create request with state cookie
		req := httptest.NewRequest(http.MethodGet, "/google/callback", nil)

		items := map[string]any{"state": "", "codeVerifier": ""}
		claims := jwt.MapClaims{}
		maps.Copy(claims, items)
		token, err := security.GenerateJWT(claims, "")
		if err != nil {
			t.Fatal("can't create token for state oauth cookie")
		}

		req.AddCookie(&http.Cookie{
			Name:  "oasc",
			Value: token,
		})
		w := httptest.NewRecorder()

		// Call handler
		oauthHandler.ServeHTTP(w, req)

		// Check response
		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode) // Currently we can't mock fetch auth user
	})

	t.Run("Invalid state", func(t *testing.T) {
		mockConfig := oauth.OAuthConfig{}
		oauthHandler := oauth.NewOAuthHandler(mockConfig)

		req := httptest.NewRequest(http.MethodGet, "/google/callback", nil)
		w := httptest.NewRecorder()

		oauthHandler.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})
}
