package main

import (
	"flag"

	"github.com/escalopa/chatterly/internal/app"
	"github.com/escalopa/chatterly/internal/auth"
	"github.com/escalopa/chatterly/internal/config"
	"github.com/escalopa/chatterly/internal/db"
	"github.com/escalopa/chatterly/internal/log"
	"github.com/escalopa/chatterly/internal/service"
)

var configPath = flag.String("config", "config.yml", "path to config file")

func main() {
	ctx := app.NewContext()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatal("load config", log.Err(err))
	}

	database, err := db.New(ctx, cfg.DB.URI)
	if err != nil {
		log.Fatal("init database", log.Err(err))
	}
	defer func() { database.Close(ctx) }()

	userTokenProvider := auth.NewUserProvider(cfg.JWT.User)
	chatTokenProvider := auth.NewChatProvider(cfg.JWT.Chat)
	oauthProvider := auth.NewOAuthProvider(cfg.OAuth)

	s := app.New(
		app.Config{
			Domain:          cfg.App.Domain,
			AllowOrigins:    cfg.App.AllowOrigins,
			AccessTokenTTL:  cfg.JWT.User.AccessTokenTTL,
			RefreshTokenTTL: cfg.JWT.User.RefreshTokenTTL,
		},
		service.New(
			database,
			oauthProvider,
			userTokenProvider,
			chatTokenProvider,
		),
	)

	err = s.Run(ctx, cfg.App.Addr)
	if err != nil {
		log.Fatal("server start", log.Err(err))
	}
}
