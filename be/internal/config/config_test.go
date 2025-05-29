package config

import (
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	configData := []byte(`
app:
  addr: ":8080"
  domain: "http://localhost:8080/api/health"
  allow_origins:
    - "http://localhost"

db:
  file: "./database.db"

jwt:
  room:
    secret_key: "your_room_secret_key"
    token_ttl: 1m
  auth:
    secret_key: "your_auth_secret_key"
    access_token_ttl: 1h
    refresh_token_ttl: 720h

oauth:
  google:
    client_id: "your-google-client-id"
    client_secret: "your-google-client-secret"
    redirect_url: "http://localhost:8080/api/oauth/google/callback"
    user_endpoint: "https://www.googleapis.com/oauth2/v2/userinfo"
  github:
    client_id: "your-github-client-id"
    client_secret: "your-github-client-secret"
    redirect_url: "http://localhost:8080/api/oauth/github/callback"
    user_endpoint: "https://api.github.com/user"
  yandex:
    client_id: "your-yandex-client-id"
    client_secret: "your-yandex-client-secret"
    redirect_url: "http://localhost:8080/api/oauth/yandex/callback"
    user_endpoint: "https://login.yandex.ru/info?format=json"
`)

	tmpFile, err := os.CreateTemp("/tmp", "config*.yml")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Remove(tmpFile.Name()))
	}()

	_, err = tmpFile.Write(configData)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	config, err := LoadConfig(tmpFile.Name())
	require.NoError(t, err)

	expectedConfig := Config{
		App: AppConfig{
			Addr:         ":8080",
			Domain:       "http://localhost:8080/api/health",
			AllowOrigins: []string{"http://localhost"},
		},
		DB: DBConfig{
			File: "./database.db",
		},
		JWT: JWTConfig{
			Room: JWTRoom{
				SecretKey: "your_room_secret_key",
				TokenTTL:  1 * time.Minute,
			},
			User: JWTUser{
				SecretKey:       "your_auth_secret_key",
				AccessTokenTTL:  1 * time.Hour,
				RefreshTokenTTL: 720 * time.Hour,
			},
		},
		OAuth: OAuthConfig{
			Google: OAuthProviderConfig{
				ClientID:     "your-google-client-id",
				ClientSecret: "your-google-client-secret",
				RedirectURL:  "http://localhost:8080/api/oauth/google/callback",
				UserEndpoint: "https://www.googleapis.com/oauth2/v2/userinfo",
			},
			GitHub: OAuthProviderConfig{
				ClientID:     "your-github-client-id",
				ClientSecret: "your-github-client-secret",
				RedirectURL:  "http://localhost:8080/api/oauth/github/callback",
				UserEndpoint: "https://api.github.com/user",
			},
			Yandex: OAuthProviderConfig{
				ClientID:     "your-yandex-client-id",
				ClientSecret: "your-yandex-client-secret",
				RedirectURL:  "http://localhost:8080/api/oauth/yandex/callback",
				UserEndpoint: "https://login.yandex.ru/info?format=json",
			},
		},
	}

	require.Empty(t, cmp.Diff(expectedConfig, config))
}
