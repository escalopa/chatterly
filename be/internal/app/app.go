package app

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/escalopa/chatterly/internal/domain"
	"github.com/escalopa/chatterly/internal/log"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type service interface {
	GetOAuthRedirectURL(provider string) (string, error)
	RegisterUser(ctx context.Context, provider string, code string) (*domain.Token, error)
	AuthenticateUser(ctx context.Context, token *domain.Token) (*domain.User, *domain.Token, error)
}

type Config struct {
	Domain       string
	AllowOrigins []string

	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration

	ShutdownTimeout time.Duration
}

type App struct {
	cfg Config

	r   *gin.Engine
	srv service
	upg *websocket.Upgrader
}

func New(cfg Config, srv service) *App {
	kors := cors.New(cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})

	upgrader := &websocket.Upgrader{
		ReadBufferSize:  1024 * 1024, // 1MB
		WriteBufferSize: 1024 * 1024, // 1MB
		CheckOrigin: func(r *http.Request) bool {
			return slices.Contains(cfg.AllowOrigins, r.Header.Get("Origin"))
		},
	}

	a := &App{
		r:   gin.Default(),
		cfg: cfg,
		srv: srv,
		upg: upgrader,
	}

	a.r.Use(kors)
	a.setup()

	return a
}

func (a *App) Run(ctx context.Context, address string) error {
	return a.run(ctx, address)
}

func (a *App) run(ctx context.Context, address string) error {
	server := &http.Server{
		Addr:    address,
		Handler: a.r,
	}

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(NewContext(), a.cfg.ShutdownTimeout)
		defer cancel()

		log.Warn("shutting down server")
		if err := server.Shutdown(ctx); err != nil {
			log.Error("server shutdown error", log.Err(err))
		}
	}()

	err := server.ListenAndServe()
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Warn("server gracefully stopped")
			return nil
		}
		return err
	}

	return nil
}

func (a *App) setup() {
	a.r.GET("/api/health", a.health)

	userRoutes := a.r.Group("/api/user")
	userRoutes.Use(a.authMiddleware)
	{
		userRoutes.GET("/info", a.getUserInfo)
		userRoutes.POST("/logout", a.logout)
	}

	//roomRoutes := a.r.Group("/api/room")
	//roomRoutes.Use(a.authMiddleware)
	//{
	//	roomRoutes.POST("/join/:room_id", a.joinRoom)
	//	roomRoutes.GET("/ws/:room_id", a.ws)
	//}

	oauthRoutes := a.r.Group("/api/oauth")
	{
		oauthRoutes.GET("/:provider", a.oauthRedirect)
		oauthRoutes.POST("/:provider/callback", a.oauthCallback)
	}
}

func (a *App) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (a *App) getUserInfo(c *gin.Context) {
	user, _ := c.Get("user")
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (a *App) logout(c *gin.Context) {
	a.setTokenCookie(c, nil)
	c.JSON(http.StatusOK, gin.H{"message": "user logged out"})
}

func (a *App) oauthRedirect(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty oauth provider"})
		return
	}

	url, err := a.srv.GetOAuthRedirectURL(provider)
	if err != nil {
		if errors.Is(err, domain.ErrOAuthUnsupportedProvider) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported oauth provider"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "temporary cannot redirect to oauth provider"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

type oauthCallbackBody struct {
	Code string `json:"code"`
}

func (a *App) oauthCallback(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty oauth provider"})
		return
	}

	var body oauthCallbackBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "corrupted request body"})
		return
	}

	if body.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty code"})
		return
	}

	token, err := a.srv.RegisterUser(c.Request.Context(), provider, body.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "temporary cannot register user"})
		return
	}

	a.setTokenCookie(c, token)
}

func (a *App) user(c *gin.Context) *domain.User {
	data, _ := c.Get("user")
	return data.(*domain.User)
}
