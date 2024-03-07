package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel string `env:"LOG_LEVEL" env-default:"debug"`

	CHBindAddr  string `env:"CH_BINDADDR" env-default:"localhost:9000"`
	CHUsername  string `env:"CH_USERNAME" env-default:"default"`
	CHDatabase  string `env:"CH_DATABASE" env-default:"default"`
	CHPassword  string `env:"CH_PASSWORD" env-default:""`
	CHBatchSize int    `env:"CH_BATCH_SIZE" env-default:"3"`

	NatsHost string `env:"NATS_BINDADDR" env-default:"localhost"`
	NatsPort string `env:"NATS_HOST" env-default:"4222"`
}

func New() *Config {
	cfg := Config{}

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		panic("error getting config")
	}

	return &cfg
}
