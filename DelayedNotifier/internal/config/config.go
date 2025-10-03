package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/wb-go/wbf/config"
)

type Config struct {
	Env       string `env:"ENV"`
	Http      HttpConfig
	Redis     RedisConfig
	Postgres  DBConfig
	RabbitMQ  RabbitMQConfig
	TelegBot  TelegramBotConfig
	EmailSmpt EmailChannel
}

type HttpConfig struct {
	Port            string        `env:"HTTP_PORT"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT"`
	ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT"`
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

type RabbitMQConfig struct {
	Host     string `env:"RABBIT_HOST"`
	Port     int    `env:"RABBIT_PORT"`
	User     string `env:"RABBIT_USER"`
	Password string `env:"RABBIT_PASSWORD"`
}

type DBConfig struct {
	Master PostgresConfig
	Slaves []PostgresConfig

	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type EmailChannel struct {
	SmptPort     int    `env:"SMPT_PORT"`
	SmptServer   string `env:"SMPT_SERVER"`
	SmptEmail    string `env:"SMPT_EMAIL"`
	SmptPassword string `env:"SMPT_PASSWORD"`
}
type TelegramBotConfig struct {
	Key    string `env:"TELEGRAMBOT_KEY"`
	ChatID int64  `env:"TELEGRAMBOT_CHATID"`
}

func LoadConfig(path string) (*Config, error) {
	wbfCfg := config.New()

	err := wbfCfg.Load(path)
	if err != nil {
		log.Fatal("could not read config file: ", err)
	}

	var cfg Config
	if err := wbfCfg.Unmarshal(&cfg); err != nil {
		log.Fatal("could not parse config file: ", err)
	}

	err = godotenv.Load(".env")
	if err != nil {
		log.Fatal("could not load .env file: ", err)
	}
	value, _ := os.LookupEnv("DB_PASSWORD")
	cfg.Postgres.Master.Password = value
	return &cfg, nil
}
