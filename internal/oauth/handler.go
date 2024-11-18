package oauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	auth "github.com/hvuhsg/gatego/internal/oauth/providers"
	"github.com/hvuhsg/gatego/pkg/security"
	"github.com/hvuhsg/gatego/pkg/session"
	"golang.org/x/oauth2"
)

const flowStateCookieName = "oasc"
const UserAuthCookieName = "uauth"

type OAuthHandler struct {
	securityKey string
	config      OAuthConfig
	muxer       *http.ServeMux
}

func NewOAuthHandler(config OAuthConfig) *OAuthHandler {
	muxer := http.NewServeMux()

	handler := &OAuthHandler{config: config, muxer: muxer}

	muxer.HandleFunc("/callback", handler.callbackHandler)
	muxer.HandleFunc("/login", handler.loginHandler)

	return handler
}

func (oa OAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	oa.muxer.ServeHTTP(w, r)
}

func (oa OAuthHandler) loginHandler(w http.ResponseWriter, r *http.Request) {
	providerName := r.URL.Query().Get("provider")

	// Get provider
	provider, err := auth.NewProviderByName(providerName)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Configure and validate provider is enabled
	err = oa.config.NamedAuthProviderConfigs()[providerName].SetupProvider(provider)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create authUrl options
	urlOpts := []oauth2.AuthCodeOption{}

	if providerName == auth.NameApple {
		// see https://developer.apple.com/documentation/sign_in_with_apple/sign_in_with_apple_js/incorporating_sign_in_with_apple_into_other_platforms#3332113
		urlOpts = append(urlOpts, oauth2.SetAuthURLParam("response_mode", "query"))
	}

	// Create challenge for PKCE supporting config
	var codeVerifier string
	if provider.PKCE() {
		codeVerifier = security.GenerateRandomString(43)
		codeChallenge := security.S256Challenge(codeVerifier)
		codeChallengeMethod := "S256"

		urlOpts = append(urlOpts,
			oauth2.SetAuthURLParam("code_challenge", codeChallenge),
			oauth2.SetAuthURLParam("code_challenge_method", codeChallengeMethod),
		)
	}

	// Add the redirect url
	urlOpts = append(urlOpts,
		oauth2.SetAuthURLParam("redirect_uri", buildRedirectURL(r.Host, oa.config.BaseURL, providerName)),
	)

	// Create random unique state
	state := security.GenerateRandomString(30)

	// Build auth url
	authUrl := provider.BuildAuthUrl(
		state,
		urlOpts...,
	)

	// Save auth session
	items := map[string]any{"state": state, "codeVerifier": codeVerifier}
	sessionCookie := session.JWTCookie{
		Cookie: &http.Cookie{Name: flowStateCookieName},
	}
	sessionCookie.SetItems(oa.securityKey, items)
	http.SetCookie(w, sessionCookie.Cookie)

	http.Redirect(w, r, authUrl, http.StatusTemporaryRedirect)
}

func (oa OAuthHandler) callbackHandler(w http.ResponseWriter, r *http.Request) {
	providerName := r.URL.Query().Get("provider")
	code := r.URL.Query().Get("code")
	stateFromProvider := r.URL.Query().Get("state")
	cookie, err := r.Cookie(flowStateCookieName)
	if err != nil {
		http.Error(w, "invalid state", http.StatusConflict)
		return
	}

	cookieSession := session.JWTCookie{Cookie: cookie}
	items, err := cookieSession.GetItems(oa.securityKey)

	if err != nil {
		http.Error(w, "invalid state", http.StatusConflict)
		return
	}

	state := items["state"].(string)
	codeVerifier := items["codeVerifier"].(string)

	// Validate state match
	if stateFromProvider != state {
		http.Error(w, "invalid state", http.StatusConflict)
		return
	}

	// Get provider by name
	provider, err := auth.NewProviderByName(providerName)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Set timeout context for provider
	context, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	provider.SetContext(context)

	// Setup provider
	providerConfig := oa.config.NamedAuthProviderConfigs()[providerName]
	err = providerConfig.SetupProvider(provider)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	provider.SetRedirectUrl(buildRedirectURL(r.Host, oa.config.BaseURL, providerName))

	var opts []oauth2.AuthCodeOption

	if provider.PKCE() {
		opts = append(opts, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	}

	// fetch token
	token, err := provider.FetchToken(code, opts...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// fetch auth user
	authUser, err := provider.FetchAuthUser(token)
	if err != nil {
		http.Error(w, "can't fetch auth user", http.StatusBadRequest)
		return
	}

	sessionCookie := &session.JWTCookie{
		Cookie: &http.Cookie{
			Name: UserAuthCookieName,
		},
	}

	err = sessionCookie.SetItems(oa.securityKey, authUser.RawUser)
	if err != nil {
		http.Error(w, "can't generate jwt token from user auth data", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, sessionCookie.Cookie)

	// Return ok message
	http.Redirect(w, r, oa.config.AfterLoginRedirect, http.StatusTemporaryRedirect)
}

// <baseURL>/provider/{provider}
func BuildOAuthEndpointURL(baseURL string) string {
	if baseURL == "" {
		baseURL = DEFAULT_BASE_URL
	}

	endpUrl, err := url.JoinPath(baseURL, "/provider/{provider}") // TODO: Check how to set path param placeholder
	if err != nil {
		panic(errors.New("can't create base oauth endpoint path url"))
	}
	return endpUrl
}

// Should create a callback url for the provider response
// Return's http://<host>/<BaseURL>/callback/<providerName>
func buildRedirectURL(host string, baseURL string, providerName string) string {
	callbackURL := BuildOAuthEndpointURL(baseURL)
	return fmt.Sprintf("http://%s/%s", host, strings.ReplaceAll(callbackURL, "{provider}", providerName))
}
