package domain

import "errors"

var (
	ErrTokenInvalid = errors.New("token invalid")
	ErrTokenExpired = errors.New("token expired")
)

var (
	ErrOAuthUnsupportedProvider = errors.New("unsupported oauth provider")
	ErrOAuthExchange            = errors.New("oauth exchange error")
	ErrOAuthGetUserInfo         = errors.New("oauth get user info error")
)

var (
	ErrDBUserNotFound = errors.New("user not found")
	ErrDBQuery        = errors.New("database query error")
)
