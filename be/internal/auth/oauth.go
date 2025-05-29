package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/escalopa/chatterly/internal/config"
	"github.com/escalopa/chatterly/internal/domain"
	"github.com/escalopa/chatterly/internal/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/gitlab"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/yandex"
)

type OAuthProvider struct {
	providers map[string]*provider
}

func NewOAuthProvider(cfg config.OAuthConfig) *OAuthProvider {
	op := &OAuthProvider{providers: make(map[string]*provider)}

	var oauthProviderConfig config.OAuthProviderConfig

	oauthProviderConfig = cfg[googleProvider]
	op.providers[googleProvider] = &provider{
		config: &oauth2.Config{
			Scopes:       oauthProviderConfig.Scopes,
			ClientID:     oauthProviderConfig.ClientID,
			ClientSecret: oauthProviderConfig.ClientSecret,
			RedirectURL:  oauthProviderConfig.RedirectURL,
			Endpoint:     google.Endpoint,
		},
		endpoint: oauthProviderConfig.UserEndpoint,
		payload:  func() payload { return &googlePayload{} },
	}

	oauthProviderConfig = cfg[yandexProvider]
	op.providers[yandexProvider] = &provider{
		config: &oauth2.Config{
			Scopes:       oauthProviderConfig.Scopes,
			ClientID:     oauthProviderConfig.ClientID,
			ClientSecret: oauthProviderConfig.ClientSecret,
			RedirectURL:  oauthProviderConfig.RedirectURL,
			Endpoint:     yandex.Endpoint,
		},
		endpoint: oauthProviderConfig.UserEndpoint,
		payload:  func() payload { return &yandexPayload{} },
	}

	oauthProviderConfig = cfg[gitlabProvider]
	op.providers[gitlabProvider] = &provider{
		config: &oauth2.Config{
			Scopes:       oauthProviderConfig.Scopes,
			ClientID:     oauthProviderConfig.ClientID,
			ClientSecret: oauthProviderConfig.ClientSecret,
			RedirectURL:  oauthProviderConfig.RedirectURL,
			Endpoint:     gitlab.Endpoint,
		},
		endpoint: oauthProviderConfig.UserEndpoint,
		payload:  func() payload { return &gitlabPayload{} },
	}

	oauthProviderConfig = cfg[githubProvider]
	op.providers[githubProvider] = &provider{
		config: &oauth2.Config{
			Scopes:       oauthProviderConfig.Scopes,
			ClientID:     oauthProviderConfig.ClientID,
			ClientSecret: oauthProviderConfig.ClientSecret,
			RedirectURL:  oauthProviderConfig.RedirectURL,
			Endpoint:     github.Endpoint,
		},
		endpoint: oauthProviderConfig.UserEndpoint,
		payload:  func() payload { return &githubPayload{} },
	}
	return op
}

func (op *OAuthProvider) GetRedirectURL(provider string) (string, error) {
	p, exists := op.providers[provider]
	if !exists {
		return "", domain.ErrOAuthUnsupportedProvider
	}

	return p.config.AuthCodeURL(provider), nil
}

func (op *OAuthProvider) HandleCallback(ctx context.Context, provider string, code string) (*domain.User, error) {
	p, exists := op.providers[provider]
	if !exists {
		return nil, domain.ErrOAuthUnsupportedProvider
	}

	// get unique token for user's data retrieval
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		log.Error("oauthConfig.Exchange", log.Err(err))
		return nil, domain.ErrOAuthExchange
	}

	// fetch user's data
	client := p.config.Client(ctx, token)
	resp, err := client.Get(p.endpoint)
	if err != nil {
		return nil, err
	}
	defer func(b io.ReadCloser) { _ = b.Close() }(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch user info: %s", resp.Status)
	}

	// parse user's data
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	dst := p.payload()
	err = json.Unmarshal(body, dst)
	if err != nil {
		return nil, err
	}

	return dst.ToUser(), nil
}
