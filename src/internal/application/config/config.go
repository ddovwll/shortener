package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel string
	Postgres PostgresConfig
	HTTP     HTTPConfig
	Kafka    KafkaConfig
}

type PostgresConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type KafkaConfig struct {
	Broker  string
	Topic   string
	GroupID string
}

type HTTPConfig struct {
	Port string
}

func Load() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return &cfg, nil
}
