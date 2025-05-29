package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/escalopa/chatterly/internal/config"
	"github.com/escalopa/chatterly/internal/domain"
	"github.com/golang-jwt/jwt/v4"
)

type (
	ChatProvider struct {
		secretKey []byte
		tokenTTL  time.Duration
	}

	chatClaims struct {
		UserID string `json:"user_id"`
		jwt.RegisteredClaims
	}
)

func NewChatProvider(cfg config.JWTChat) *ChatProvider {
	return &ChatProvider{
		secretKey: []byte(cfg.SecretKey),
		tokenTTL:  cfg.TokenTTL,
	}
}

func (cp *ChatProvider) CreateToken(userID string, sessionID string) (string, error) {
	now := time.Now()
	claims := chatClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(cp.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(cp.secretKey)
}

func (cp *ChatProvider) VerifyToken(tokenStr string) (*domain.ChatTokenPayload, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &chatClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return cp.secretKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domain.ErrTokenExpired
		}
		return nil, domain.ErrTokenInvalid
	}

	claims, ok := token.Claims.(*chatClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrTokenInvalid
	}

	p := &domain.ChatTokenPayload{UserID: claims.UserID}

	return p, nil
}

type (
	UserProvider struct {
		secretKey       []byte
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
	}

	userClaims struct {
		UserID    string `json:"user_id"`
		Email     string `json:"email"`
		SessionID string `json:"session_id"`
		jwt.RegisteredClaims
	}
)

func NewUserProvider(cfg config.JWTUser) *UserProvider {
	return &UserProvider{
		secretKey:       []byte(cfg.SecretKey),
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
	}
}

func (up *UserProvider) CreateToken(userID string, email string) (*domain.Token, error) {
	now := time.Now()

	accessToken, err := up.createToken(userID, email, up.accessTokenTTL, now)
	if err != nil {
		return nil, err
	}

	refreshToken, err := up.createToken(userID, email, up.refreshTokenTTL, now)
	if err != nil {
		return nil, err
	}

	token := &domain.Token{
		Access:  accessToken,
		Refresh: refreshToken,
	}

	return token, nil
}

func (up *UserProvider) createToken(userID string, email string, ttl time.Duration, now time.Time) (string, error) {
	claims := userClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(up.secretKey)
}

func (up *UserProvider) VerifyToken(tokenStr string) (*domain.UserTokenPayload, error) {
	if tokenStr == "" {
		return nil, domain.ErrTokenExpired // treat empty token as expired
	}

	token, err := jwt.ParseWithClaims(tokenStr, &userClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return up.secretKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domain.ErrTokenExpired
		}
		return nil, domain.ErrTokenInvalid
	}

	claims, ok := token.Claims.(*userClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrTokenInvalid
	}

	p := &domain.UserTokenPayload{
		UserID: claims.UserID,
		Email:  claims.Email,
	}

	return p, nil
}
