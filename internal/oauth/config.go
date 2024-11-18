package oauth

import (
	"errors"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	auth "github.com/hvuhsg/gatego/internal/oauth/providers"
)

const DEFAULT_BASE_URL = "/oauth"

type OAuthConfig struct {
	BaseURL            string `yaml:"base-url"`
	AfterLoginRedirect string `yaml:"after-login-redirect-to"`

	// Auth providers
	Google    AuthProviderConfig `yaml:"google"`
	Facebook  AuthProviderConfig `yaml:"facebook"`
	Github    AuthProviderConfig `yaml:"github"`
	Gitlab    AuthProviderConfig `yaml:"gitlab"`
	Discord   AuthProviderConfig `yaml:"discord"`
	Twitter   AuthProviderConfig `yaml:"twitter"`
	Microsoft AuthProviderConfig `yaml:"microsoft"`
	Spotify   AuthProviderConfig `yaml:"spotify"`
	Kakao     AuthProviderConfig `yaml:"kakao"`
	Twitch    AuthProviderConfig `yaml:"twitch"`
	Strava    AuthProviderConfig `yaml:"strava"`
	Gitee     AuthProviderConfig `yaml:"gitee"`
	Livechat  AuthProviderConfig `yaml:"livechat"`
	Gitea     AuthProviderConfig `yaml:"gitea"`
	OIDC      AuthProviderConfig `yaml:"oidc"`
	OIDC2     AuthProviderConfig `yaml:"oidc2"`
	OIDC3     AuthProviderConfig `yaml:"oidc3"`
	Apple     AuthProviderConfig `yaml:"apple"`
	Instagram AuthProviderConfig `yaml:"instagram"`
	VK        AuthProviderConfig `yaml:"VK"`
	Yandex    AuthProviderConfig `yaml:"yandex"`
	Patreon   AuthProviderConfig `yaml:"patreon"`
	Mailcow   AuthProviderConfig `yaml:"mailcow"`
	Bitbucket AuthProviderConfig `yaml:"bitbucket"`
}

func (o OAuthConfig) Validate() error {
	if err := o.Google.Validate(); err != nil {
		return err
	}
	if err := o.Facebook.Validate(); err != nil {
		return err
	}
	if err := o.Github.Validate(); err != nil {
		return err
	}
	if err := o.Gitlab.Validate(); err != nil {
		return err
	}
	if err := o.Discord.Validate(); err != nil {
		return err
	}
	if err := o.Twitter.Validate(); err != nil {
		return err
	}
	if err := o.Microsoft.Validate(); err != nil {
		return err
	}
	if err := o.Spotify.Validate(); err != nil {
		return err
	}
	if err := o.Kakao.Validate(); err != nil {
		return err
	}
	if err := o.Twitch.Validate(); err != nil {
		return err
	}
	if err := o.Strava.Validate(); err != nil {
		return err
	}
	if err := o.Gitee.Validate(); err != nil {
		return err
	}
	if err := o.Livechat.Validate(); err != nil {
		return err
	}
	if err := o.Gitea.Validate(); err != nil {
		return err
	}
	if err := o.OIDC.Validate(); err != nil {
		return err
	}
	if err := o.OIDC2.Validate(); err != nil {
		return err
	}
	if err := o.OIDC3.Validate(); err != nil {
		return err
	}
	if err := o.Apple.Validate(); err != nil {
		return err
	}
	if err := o.Instagram.Validate(); err != nil {
		return err
	}
	if err := o.VK.Validate(); err != nil {
		return err
	}
	if err := o.Yandex.Validate(); err != nil {
		return err
	}
	if err := o.Patreon.Validate(); err != nil {
		return err
	}
	if err := o.Mailcow.Validate(); err != nil {
		return err
	}
	if err := o.Bitbucket.Validate(); err != nil {
		return err
	}

	return nil
}

// NamedAuthProviderConfigs returns a map with all registered OAuth2
// provider configurations (indexed by their name identifier).
func (s OAuthConfig) NamedAuthProviderConfigs() map[string]AuthProviderConfig {
	return map[string]AuthProviderConfig{
		auth.NameGoogle:     s.Google,
		auth.NameFacebook:   s.Facebook,
		auth.NameGithub:     s.Github,
		auth.NameGitlab:     s.Gitlab,
		auth.NameDiscord:    s.Discord,
		auth.NameTwitter:    s.Twitter,
		auth.NameMicrosoft:  s.Microsoft,
		auth.NameSpotify:    s.Spotify,
		auth.NameKakao:      s.Kakao,
		auth.NameTwitch:     s.Twitch,
		auth.NameStrava:     s.Strava,
		auth.NameGitee:      s.Gitee,
		auth.NameLivechat:   s.Livechat,
		auth.NameGitea:      s.Gitea,
		auth.NameOIDC:       s.OIDC,
		auth.NameOIDC + "2": s.OIDC2,
		auth.NameOIDC + "3": s.OIDC3,
		auth.NameApple:      s.Apple,
		auth.NameInstagram:  s.Instagram,
		auth.NameVK:         s.VK,
		auth.NameYandex:     s.Yandex,
		auth.NamePatreon:    s.Patreon,
		auth.NameMailcow:    s.Mailcow,
		auth.NameBitbucket:  s.Bitbucket,
	}
}

type AuthProviderConfig struct {
	Enabled      bool   `form:"enabled" yaml:"enabled"`
	ClientId     string `form:"clientId" yaml:"clientId"`
	ClientSecret string `form:"clientSecret" yaml:"clientSecret"`
	AuthUrl      string `form:"authUrl" yaml:"authUrl"`
	TokenUrl     string `form:"tokenUrl" yaml:"tokenUrl"`
	UserApiUrl   string `form:"userApiUrl" yaml:"userApiUrl"`
	DisplayName  string `form:"displayName" yaml:"displayName"`
	PKCE         *bool  `form:"pkce" yaml:"pkce"`
}

func (c AuthProviderConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.ClientId, validation.When(c.Enabled, validation.Required)),
		validation.Field(&c.ClientSecret, validation.When(c.Enabled, validation.Required)),
		validation.Field(&c.AuthUrl, is.URL),
		validation.Field(&c.TokenUrl, is.URL),
		validation.Field(&c.UserApiUrl, is.URL),
	)
}

// SetupProvider loads the current AuthProviderConfig into the specified provider.
func (c AuthProviderConfig) SetupProvider(provider auth.Provider) error {
	if !c.Enabled {
		return errors.New("the provider is not enabled")
	}

	if c.ClientId != "" {
		provider.SetClientId(c.ClientId)
	}

	if c.ClientSecret != "" {
		provider.SetClientSecret(c.ClientSecret)
	}

	if c.AuthUrl != "" {
		provider.SetAuthUrl(c.AuthUrl)
	}

	if c.UserApiUrl != "" {
		provider.SetUserApiUrl(c.UserApiUrl)
	}

	if c.TokenUrl != "" {
		provider.SetTokenUrl(c.TokenUrl)
	}

	if c.DisplayName != "" {
		provider.SetDisplayName(c.DisplayName)
	}

	if c.PKCE != nil {
		provider.SetPKCE(*c.PKCE)
	}

	return nil
}
