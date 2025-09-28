package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
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

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, continuing with system environment variables")
	}

	cfg := Config{}

	cfg.Env = os.Getenv("ENV")
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

	cfg.Redis.Addr = os.Getenv("REDIS_ADDR")

	cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")
	cfg.Redis.DBRedis, _ = strconv.Atoi(os.Getenv("REDIS_DBREDIS"))

	cfg.RabbitMQ.Host = os.Getenv("RABBIT_HOST")
	cfg.RabbitMQ.Port, _ = strconv.Atoi(os.Getenv("RABBIT_PORT"))
	cfg.RabbitMQ.User = os.Getenv("RABBIT_USER")
	cfg.RabbitMQ.Password = os.Getenv("RABBIT_PASSWORD")

	cfg.TelegBot.Key = os.Getenv("TELEGRAMBOT_KEY")
	cfg.TelegBot.ChatID, _ = strconv.ParseInt(os.Getenv("TELEGRAMBOT_CHATID"), 10, 64)

	cfg.EmailSmpt.SmptPort, _ = strconv.Atoi(os.Getenv("SMPT_PORT"))
	cfg.EmailSmpt.SmptPassword = os.Getenv("SMPT_PASSWORD")
	cfg.EmailSmpt.SmptEmail = os.Getenv("SMPT_EMAIL")
	cfg.EmailSmpt.SmptServer = os.Getenv("SMPT_SERVER")
	// Отладочный вывод всех значений
	log.Printf("Config loaded: %+v\n", cfg)

	return &cfg, nil
}

/*
	package config

import (

	"fmt"
	"time"









	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/zlog"

)

	type Config struct {
		Env       string            `mapstructure:"env"`
		Http      HttpConfig        `mapstructure:"http"`
		Redis     RedisConfig       `mapstructure:"redis"`
		Postgres  PostgresConfig    `mapstructure:"postgres"`
		RabbitMQ  RabbitMQConfig    `mapstructure:"rabbitmq"`
		TelegBot  TelegramBotConfig `mapstructure:"telegbot"`
		EmailSmpt EmailChannel      `mapstructure:"email_smpt"`
	}

	type HttpConfig struct {
		Port            string        `mapstructure:"port"`
		ReadTimeout     time.Duration `mapstructure:"read_timeout"`
		WriteTimeout    time.Duration `mapstructure:"write_timeout"`
		ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	}

	type PostgresConfig struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Database string `mapstructure:"database"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		SSLMode  string `mapstructure:"ssl_mode"`
	}

	type RedisConfig struct {
		Addr     string `mapstructure:"addr"`
		Password string `mapstructure:"password"`
		DBRedis  int    `mapstructure:"db_redis"`
	}

	type RabbitMQConfig struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
	}

	type EmailChannel struct {
		SmptPort     int    `mapstructure:"smpt_port"`
		SmptServer   string `mapstructure:"smpt_server"`
		SmptEmail    string `mapstructure:"smpt_email"`
		SmptPassword string `mapstructure:"smpt_password"`
	}

	type TelegramBotConfig struct {
		Key    string `mapstructure:"key"`
		ChatID int64  `mapstructure:"chat_id"`
	}

// LoadConfig loads configuration from file and environment variables using viper

	func LoadConfig() (*Config, error) {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./config")
		viper.AutomaticEnv()

		// Bind environment variables explicitly if needed (example below)
		bindings := map[string]string{
			"env":                      "ENV",
			"http.port":                "HTTP_PORT",
			"http.read_timeout":        "HTTP_READ_TIMEOUT",
			"http.write_timeout":       "HTTP_WRITE_TIMEOUT",
			"http.shutdown_timeout":    "HTTP_SHUTDOWN_TIMEOUT",
			"postgres.host":            "POSTGRES_HOST",
			"postgres.port":            "POSTGRES_PORT",
			"postgres.database":        "POSTGRES_DATABASE",
			"postgres.user":            "POSTGRES_USER",
			"postgres.password":        "POSTGRES_PASSWORD",
			"postgres.ssl_mode":        "POSTGRES_SSL_MODE",
			"redis.addr":               "REDIS_ADDR",
			"redis.password":           "REDIS_PASSWORD",
			"redis.db_redis":           "REDIS_DBREDIS",
			"rabbitmq.host":            "RABBIT_HOST",
			"rabbitmq.port":            "RABBIT_PORT",
			"rabbitmq.user":            "RABBIT_USER",
			"rabbitmq.password":        "RABBIT_PASSWORD",
			"telegbot.key":             "TELEGRAMBOT_KEY",
			"telegbot.chat_id":         "TELEGRAMBOT_CHATID",
			"email_smpt.smpt_port":     "SMPT_PORT",
			"email_smpt.smpt_password": "SMPT_PASSWORD",
			"email_smpt.smpt_email":    "SMPT_EMAIL",
			"email_smpt.smpt_server":   "SMPT_SERVER",
		}
		for key, env := range bindings {
			if err := viper.BindEnv(key, env); err != nil {
				return nil, fmt.Errorf("failed to bind env %s: %w", env, err)
			}
		}

		if err := viper.ReadInConfig(); err != nil {
			// log but continue if file not found, depends on your problem
			zlog.Logger.Warn().Err(err).Msg("config file not found or unable to read config file")
		}

		var cfg Config
		if err := viper.Unmarshal(&cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}

		return &cfg, nil
	}
*/

/*
package config

import (
	"fmt"
	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/zlog"
	"time"
)

// AppConfig описывает структуру конфигурации приложения
type Config struct {
	Env       string            `mapstructure:"env"`
	Http      HttpConfig        `mapstructure:"http"`
	Redis     RedisConfig       `mapstructure:"redis"`
	Postgres  DBConfig          `mapstructure:"db"`
	RabbitMQ  RabbitMQConfig    `mapstructure:"rabbitmq"`
	TelegBot  TelegramBotConfig `mapstructure:"telegbot"`
	EmailSmpt EmailChannel      `mapstructure:"email_smpt"`
}

type HttpConfig struct {
	Port            string        `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DBRedis  int    `mapstructure:"db_redis"`
}

type RabbitMQConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

type DBConfig struct {
	Master PostgresConfig   `mapstructure:"master"`
	Slaves []PostgresConfig `mapstructure:"slaves"`

	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type EmailChannel struct {
	SmptPort     int    `mapstructure:"smpt_port"`
	SmptServer   string `mapstructure:"smpt_server"`
	SmptEmail    string `mapstructure:"smpt_email"`
	SmptPassword string `mapstructure:"smpt_password"`
}

type TelegramBotConfig struct {
	Key    string `mapstructure:"key"`
	ChatID int64  `mapstructure:"chat_id"`
}

// LoadConfig загружает конфигурацию из файла config.yaml и окружения
func LoadConfig(path string) (*Config, error) {
	cfg := config.New() // создание обёртки из github.com/wb-go/wbf/config

	// Значения по умолчанию
	cfg.SetDefault("http.read_timeout", "10s")
	cfg.SetDefault("http.write_timeout", "10s")
	cfg.SetDefault("http.shutdown_timeout", "10s")

	if err := cfg.Load(path); err != nil {
		zlog.Logger.Warn().Err(err).Msg("Не удалось прочитать конфигурационный файл")
		// Опционально: можно возвращать ошибку, если файл обязателен
	}

	var appCfg Config
	if err := cfg.Unmarshal(&appCfg); err != nil {
		return nil, fmt.Errorf("ошибка демаршалинга конфигурации: %w", err)
	}

	return &appCfg, nil
}
*/
