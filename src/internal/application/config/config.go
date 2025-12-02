package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel string `env:"LOG_LEVEL" env-default:"info"`
	Postgres PostgresConfig
	HTTP     HTTPConfig
	Kafka    KafkaConfig
	Redis    RedisConfig
}

type PostgresConfig struct {
	DSN             string        `env:"PG_DSN" yaml:"dsn" env-default:"postgres://postgres:postgres@db:5432/postgres?sslmode=disable"`
	MaxOpenConns    int           `env:"PG_MAX_OPEN_CONNS" env-default:"50"`
	MaxIdleConns    int           `env:"PG_MAX_IDLE_CONNS" env-default:"10"`
	ConnMaxLifetime time.Duration `env:"PG_CONN_MAX_LIFETIME" env-default:"1m"`
}

type KafkaConfig struct {
	Broker  string `env:"KAFKA_BROKER" env-default:"kafka:9092"`
	Topic   string `env:"KAFKA_TOPIC" env-default:"visits"`
	GroupID string `env:"KAFKA_GROUP_ID" env-default:"visits-group"`
}

type HTTPConfig struct {
	Port string `env:"HTTP_PORT" env-default:":8080"`
}

type RedisConfig struct {
	Host     string `env:"REDIS_HOST" env-default:"redis:6379"`
	Password string `env:"REDIS_PASSWORD" env-default:""`
	DB       int    `env:"REDIS_DB" env-default:"0"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return &cfg, nil
}
