package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Http       HttpConfig
	Redis      RedisConfig
	AuthConfig AuthConfig
	Postgres   DBConfig
}

type HttpConfig struct {
	Port            string        `env:"HTTP_PORT"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT"`
	ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT"`
}

type AuthConfig struct {
	AccessTokenTTL  time.Duration `env:"ACCESS_TTL" env-default:"15m"`
	RefreshTokenTTL time.Duration `env:"REFRESH_TTL" env-default:"720h"`
	PasswordSalt    string        `env:"PASSWORD_SALT" env-required:"true"`
	JWTSigningKey   string        `env:"JWT_SIGNING_KEY" env-required:"true"`
}
type PostgresConfig struct {
	Host     string `env:"POSTGRES_HOST"`
	Port     int    `env:"POSTGRES_PORT"`
	Database string `env:"POSTGRES_DATABASE"`
	User     string `env:"POSTGRES_USER"`
	Password string `env:"POSTGRES_PASSWORD"`
	SSLMode  string `env:"POSTGRES_SSL_MODE"`
}

type RedisConfig struct {
	Addr     string `env:"REDIS_ADDR"`
	Password string `env:"REDIS_PASSWORD"`
	DBRedis  int    `env:"REDIS_DBREDIS"`
}

type DBConfig struct {
	Master PostgresConfig
	Slaves []PostgresConfig

	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or could not be loaded, continuing with system env variables: %v", err)
	}

	var cfg Config

	cfg.Http.Port = os.Getenv("HTTP_PORT")

	readTimeout := os.Getenv("HTTP_READ_TIMEOUT")
	if readTimeout == "" {
		readTimeout = "10s"
	}
	cfg.Http.ReadTimeout, _ = time.ParseDuration(readTimeout)

	writeTimeout := os.Getenv("HTTP_WRITE_TIMEOUT")
	if writeTimeout == "" {
		writeTimeout = "10s"
	}
	cfg.Http.WriteTimeout, _ = time.ParseDuration(writeTimeout)

	shutdownTimeout := os.Getenv("HTTP_SHUTDOWN_TIMEOUT")
	if shutdownTimeout == "" {
		shutdownTimeout = "10s"
	}
	cfg.Http.ShutdownTimeout, _ = time.ParseDuration(shutdownTimeout)

	cfg.Postgres.Master.Host = os.Getenv("POSTGRES_HOST")
	cfg.Postgres.Master.Port, _ = strconv.Atoi(os.Getenv("POSTGRES_PORT"))
	cfg.Postgres.Master.Database = os.Getenv("POSTGRES_DATABASE")
	cfg.Postgres.Master.User = os.Getenv("POSTGRES_USER")
	cfg.Postgres.Master.Password = os.Getenv("POSTGRES_PASSWORD")
	cfg.Postgres.Master.SSLMode = os.Getenv("POSTGRES_SSL_MODE")
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
	cfg.Redis.Addr = os.Getenv("REDIS_ADDR")

	cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")
	cfg.Redis.DBRedis, _ = strconv.Atoi(os.Getenv("REDIS_DBREDIS"))

	// Отладочный вывод всех значений
	log.Printf("Config loaded: %+v\n", cfg)

	return &cfg, nil
}
