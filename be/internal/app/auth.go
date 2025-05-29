package app

import (
	"errors"
	"net/http"

	"github.com/escalopa/chatterly/internal/domain"
	"github.com/escalopa/chatterly/internal/log"
	"github.com/gin-gonic/gin"
)

const (
	accessTokenKey  = "X-Access-Token"
	refreshTokenKey = "X-Refresh-Token"
)

func (a *App) authMiddleware(c *gin.Context) {
	accessToken, err := c.Cookie(accessTokenKey)
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	refreshToken, err := c.Cookie(refreshTokenKey)
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if accessToken == "" && refreshToken == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated user"})
		return
	}

	token := &domain.Token{Access: accessToken, Refresh: refreshToken}
	user, token, err := a.srv.AuthenticateUser(c.Request.Context(), token)
	if err != nil {
		log.Error("srv.AuthenticateUser", log.Err(err))
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "internal server error"})
		return
	}

	// if tokens were refreshed then set it back in the cookies
	if token != nil {
		a.setTokenCookie(c, token)
	}

	c.Set("user", user)
	c.Next()
}

const (
	cookiePath     = "/"
	cookieSecure   = true
	cookieHttpOnly = true
)

func (a *App) setTokenCookie(c *gin.Context, token *domain.Token) {
	var (
		accessToken, refreshToken             string
		accessTokenExpiry, refreshTokenExpiry int
	)

	if token == nil { // delete the cookies
		accessTokenExpiry = -1
		refreshTokenExpiry = -1
	} else {
		accessToken = token.Access
		refreshToken = token.Refresh
		accessTokenExpiry = int(a.cfg.AccessTokenTTL.Seconds())
		refreshTokenExpiry = int(a.cfg.RefreshTokenTTL.Seconds())
	}

	// set access token cookie
	c.SetCookie(
		accessTokenKey,
		accessToken,
		accessTokenExpiry,
		cookiePath,
		a.cfg.Domain,
		cookieSecure,
		cookieHttpOnly,
	)

	// set refresh token cookie
	c.SetCookie(
		refreshTokenKey,
		refreshToken,
		refreshTokenExpiry,
		cookiePath,
		a.cfg.Domain,
		cookieSecure,
		cookieHttpOnly,
	)
}
