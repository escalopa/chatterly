package config

import (
	"path"
	"time"

	"github.com/escalopa/chatterly/internal/log"
	"github.com/spf13/viper"
)

type Config struct {
	App AppConfig `mapstructure:"APP" json:"app" yaml:"app"`
	JWT JWTConfig `mapstructure:"JWT" json:"jwt" yaml:"jwt"`

	DB     DBConfig     `mapstructure:"DB" json:"db" yaml:"db"`
	Broker BrokerConfig `mapstructure:"BROKER" json:"broker" yaml:"broker"`

	OAuth OAuthConfig `mapstructure:"OAUTH" json:"oauth" yaml:"oauth"`
}

type AppConfig struct {
	Addr            string        `mapstructure:"ADDR" json:"addr" yaml:"addr"`
	Domain          string        `mapstructure:"DOMAIN" json:"domain" yaml:"domain"`
	AllowOrigins    []string      `mapstructure:"ALLOW_ORIGINS" json:"allow_origins" yaml:"allow_origins"`
	ShutdownTimeout time.Duration `mapstructure:"SHUTDOWN_TIMEOUT" json:"shutdown_timeout" yaml:"shutdown_timeout"`
}

type JWTConfig struct {
	User JWTUser `mapstructure:"USER" json:"user" yaml:"user"`
	Chat JWTChat `mapstructure:"CHAT" json:"chat" yaml:"chat"`
}

type JWTUser struct {
	SecretKey       string        `mapstructure:"SECRET_KEY" json:"secret_key" yaml:"secret_key"`
	AccessTokenTTL  time.Duration `mapstructure:"ACCESS_TOKEN_TTL" json:"access_token_ttl" yaml:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"REFRESH_TOKEN_TTL" json:"refresh_token_ttl" yaml:"refresh_token_ttl"`
}

type JWTChat struct {
	SecretKey string        `mapstructure:"SECRET_KEY" json:"secret_key" yaml:"secret_key"`
	TokenTTL  time.Duration `mapstructure:"TOKEN_TTL" json:"token_ttl" yaml:"token_ttl"`
}

type DBConfig struct {
	URI string `mapstructure:"URI" json:"uri" yaml:"uri"`
}

type BrokerConfig struct {
	Servers []string `mapstructure:"SERVERS" json:"servers" yaml:"servers"`
}

type OAuthConfig map[string]OAuthProviderConfig

type OAuthProviderConfig struct {
	Scopes       []string `mapstructure:"SCOPES" json:"scopes" yaml:"scopes"`
	ClientID     string   `mapstructure:"CLIENT_ID" json:"client_id" yaml:"client_id"`
	ClientSecret string   `mapstructure:"CLIENT_SECRET" json:"client_secret" yaml:"client_secret"`
	RedirectURL  string   `mapstructure:"REDIRECT_URL" json:"redirect_url" yaml:"redirect_url"`
	UserEndpoint string   `mapstructure:"USER_ENDPOINT" json:"user_endpoint" yaml:"user_endpoint"`
}

func LoadConfig(file string) (*Config, error) {
	viper.SetConfigName(path.Base(file))
	viper.SetConfigType(path.Ext(file)[1:]) // remove dot
	viper.AddConfigPath(path.Dir(file))

	if err := viper.ReadInConfig(); err != nil {
		log.Error("read config file", log.Err(err))
		return nil, err
	}

	viper.AutomaticEnv()

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		log.Error("decode config into struct", log.Err(err))
		return nil, err
	}

	return config, nil
}
