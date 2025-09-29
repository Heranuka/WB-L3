package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env        string `env:"ENV"`
	Http       HttpConfig
	AuthConfig AuthConfig
	Postgres   PostgresConfig
}

type HttpConfig struct {
	Port            string        `env:"PORT"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env-default:"10s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env-default:"10s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env-default:"10s"`
}

type AuthConfig struct {
	AccessTokenTTL  time.Duration `env:"ACCESS_TTL" env-default:"15m"`
	RefreshTokenTTL time.Duration `env:"REFRESH_TTL" env-default:"720h"`
	PasswordSalt    string        `env:"PASSWORD_SALT" env-required:"true"`
	JWTSigningKey   string        `env:"JWT_SIGNING_KEY" env-required:"true"`
}
type PostgresConfig struct {
	Host     string `env:"POSTGRES_HOST"`
	Port     string `env:"POSTGRES_PORT"`
	Database string `env:"POSTGRES_DATABASE"`
	User     string `env:"POSTGRES_USER"`
	Password string `env:"POSTGRES_PASSWORD"`
	SSLMode  string `env:"POSTGRES_SSL_MODE"`
}

func LoadPath() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file")
	}

	var cfg Config

	cfg.Postgres.Host = os.Getenv("POSTGRES_HOST")
	cfg.Postgres.Password = os.Getenv("POSTGRES_PASSWORD")
	cfg.Postgres.Database = os.Getenv("POSTGRES_DATABASE")
	cfg.Postgres.SSLMode = os.Getenv("POSTGRES_SSL_MODE")
	cfg.Postgres.Port = os.Getenv("POSTGRES_PORT")
	cfg.Postgres.User = os.Getenv("POSTGRES_USER")
	cfg.Http.Port = os.Getenv("PORT")
	cfg.AuthConfig.PasswordSalt = os.Getenv("PASSWORD_SALT")
	cfg.AuthConfig.JWTSigningKey = os.Getenv("JWT_SIGNING_KEY")
	accessTTLStr := os.Getenv("ACCESS_TTL")
	refreshTTLStr := os.Getenv("REFRESH_TTL")

	accessTTL, err := time.ParseDuration(accessTTLStr)
	if err != nil {
		return nil, fmt.Errorf("invalid ACCESS_TTL: %w", err)
	}
	refreshTTL, err := time.ParseDuration(refreshTTLStr)
	if err != nil {
		return nil, fmt.Errorf("invalid REFRESH_TTL: %w", err)
	}

	cfg.AuthConfig.AccessTokenTTL = accessTTL
	cfg.AuthConfig.RefreshTokenTTL = refreshTTL

	return &cfg, nil
}
