package auth

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/escalopa/chatterly/internal/config"
	"github.com/escalopa/chatterly/internal/domain"
	"github.com/stretchr/testify/require"
)

const (
	testUserID    = "6656e4cf03c748fe2b3a3f92"
	testEmail     = "test@example.com"
	testSessionID = "05dmJUrW0NJNkLrcFhW"
)

var (
	expiredToken, _ = base64.RawStdEncoding.DecodeString("ZXlKaGJHY2lPaUpJVXpJMU5pSXNJblI1Y0NJNklrcFhWQ0o5LmV5SjFjMlZ5WDJsa0lqb2lOalkxTm1VMFkyWXdNMk0zTkRobVpUSmlNMkV6WmpreUlpd2ljMlZ6YzJsdmJsOXBaQ0k2SWpBMVpHMUtWWEpYTUU1S1RtdE1jbU5HYUZjaUxDSmxlSEFpT2pFM05EZzFOVFl3TWpnc0ltbGhkQ0k2TVRjME9EVTFOall5T0gwLmRvNThfOVhrVmNYc2RkVkVpMGZqM21YQ0d0NWZsLW9iVFMxbUVGT2ZIWjQ")
)

func TestChatProvider(t *testing.T) {
	t.Parallel()

	cfg := config.JWTChat{
		SecretKey: "test-secret",
		TokenTTL:  time.Minute,
	}

	cp := NewChatProvider(cfg)

	tests := []struct {
		name      string
		userID    string
		sessionID string
		modify    func(token string) string
		expectErr error
	}{
		{
			name:      "valid_token",
			userID:    testUserID,
			sessionID: testSessionID,
			modify:    func(token string) string { return token },
			expectErr: nil,
		},
		{
			name:      "expired_token",
			userID:    testUserID,
			sessionID: testSessionID,
			modify:    func(_ string) string { return string(expiredToken) },
			expectErr: domain.ErrTokenExpired,
		},
		{
			name:      "invalid_token",
			userID:    testUserID,
			sessionID: testSessionID,
			modify:    func(token string) string { return token + "invalid" },
			expectErr: domain.ErrTokenInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			token, err := cp.CreateToken(tt.userID, tt.sessionID)
			require.NoError(t, err)

			token = tt.modify(token)
			p, err := cp.VerifyToken(token)

			if tt.expectErr == nil {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.Equal(t, tt.userID, p.UserID)
				require.Equal(t, tt.sessionID, p.SessionID)
			} else {
				require.ErrorIs(t, err, tt.expectErr)
				require.Nil(t, p)
			}
		})
	}
}

func TestUserProvider(t *testing.T) {
	t.Parallel()

	cfg := config.JWTUser{
		SecretKey:       "test-secret",
		AccessTokenTTL:  time.Minute,
		RefreshTokenTTL: time.Hour,
	}
	up := NewUserProvider(cfg)

	tests := []struct {
		name      string
		userID    string
		email     string
		modify    func(token string) string
		expectErr error
	}{
		//{
		//	name:      "valid_access_token",
		//	userID:    testUserID,
		//	email:     testEmail,
		//	modify:    func(token string) string { return token },
		//	expectErr: nil,
		//},
		{
			name:      "expired_access_token",
			userID:    testUserID,
			email:     testEmail,
			modify:    func(_ string) string { return string(expiredToken) },
			expectErr: domain.ErrTokenExpired,
		},
		{
			name:      "invalid_access_token",
			userID:    testUserID,
			email:     testEmail,
			modify:    func(token string) string { return token + "invalid" },
			expectErr: domain.ErrTokenInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			token, err := up.CreateToken(tt.userID, tt.email)
			require.NoError(t, err)

			token.Access = tt.modify(token.Access)
			p, err := up.VerifyToken(token.Access)

			if tt.expectErr == nil {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.Equal(t, tt.userID, p.UserID)
				require.Equal(t, tt.email, p.Email)
			} else {
				require.ErrorIs(t, err, tt.expectErr)
				require.Nil(t, p)
			}
		})
	}
}
