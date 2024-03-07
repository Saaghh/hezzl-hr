package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	BindAddress string `env:"BIND_ADDR" env-default:":8080"`
	LogLevel    string `env:"LOG_LEVEL" env-default:"debug"`

	PGHost     string `env:"PG_HOST" env-default:"localhost"`
	PGPort     string `env:"PG_PORT" env-default:"5432"`
	PGDatabase string `env:"PG_DATABASE" env-default:"postgres"`
	PGUser     string `env:"PG_USER" env-default:"user"`
	PGPassword string `env:"PG_PASSWORD" env-default:"secret"`

	RedisAddr           string        `env:"REDIS_ADDR" env-default:"localhost:6379"`
	RedisPassword       string        `env:"REDIS_DB" env-default:""`
	RedisDB             int           `env:"REDIS_DB"`
	RedisDefaultTimeout time.Duration `env:"REDIS_TIMEOUT"`
}

func New() *Config {
	cfg := Config{}

	cfg.RedisDB = 0
	cfg.RedisDefaultTimeout = time.Minute

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		panic("error getting config")
	}

	return &cfg
}
