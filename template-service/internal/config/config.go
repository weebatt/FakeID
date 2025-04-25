package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPServer struct {
	Host       string `yaml:"host" env:"HOST" validate:"required"`
	Port       string `yaml:"port" env:"PORT" env-default:"8080" validate:"required,numeric"`
	MaxRetries int    `yaml:"max_retries" env:"MAX_RETRIES" env-default:"5" validate:"gte=1"`
	RetryDelay int    `yaml:"retry_delay" env:"RETRY_DELAY" env-default:"5" validate:"gte=1"`
}

type PostgresConfig struct {
	Host       string `yaml:"host" env:"PG_HOST" validate:"required"`
	Port       string `yaml:"port" env:"PG_PORT" env-default:"5432" validate:"required,numeric"`
	User       string `yaml:"user" env:"PG_USER" env-default:"postgres" validate:"required"`
	Password   string `yaml:"password" env:"PG_PASSWORD" env-default:""`
	DBName     string `yaml:"dbname" env:"PG_DBNAME" validate:"required"`
	SSLMode    string `yaml:"sslmode" env:"PG_SSLMODE" env-default:"disable" validate:"oneof=disable require"`
	MaxConns   int32  `yaml:"max_conns" env:"PG_MAX_CONNS" env-default:"50" validate:"gte=1"`
	MinConns   int32  `yaml:"min_conns" env:"PG_MIN_CONNS" env-default:"10" validate:"gte=1"`
	Timeout    int    `yaml:"timeout" env:"PG_TIMEOUT" env-default:"5" validate:"gte=1"`
	MaxRetries int    `yaml:"max_retries" env:"PG_MAX_RETRIES" env-default:"5" validate:"gte=1"`
	RetryDelay int    `yaml:"retry_delay" env:"PG_RETRY_DELAY" env-default:"2" validate:"gte=1"`
}

type RedisConfig struct {
	Host       string `yaml:"host" env:"REDIS_HOST" validate:"required"`
	Port       string `yaml:"port" env:"REDIS_PORT" env-default:"6379" validate:"required,numeric"`
	Password   string `yaml:"password" env:"REDIS_PASSWORD" env-default:""`
	DB         int    `yaml:"db" env:"REDIS_DB" env-default:"0" validate:"gte=0"`
	Timeout    int    `yaml:"timeout" env:"REDIS_TIMEOUT" env-default:"5" validate:"gte=1"`
	MaxRetries int    `yaml:"max_retries" env:"REDIS_MAX_RETRIES" env-default:"5" validate:"gte=1"`
	RetryDelay int    `yaml:"retry_delay" env:"REDIS_RETRY_DELAY" env-default:"3" validate:"gte=1"`
}

type Config struct {
	Env        string         `yaml:"env" env:"ENV" env-default:"prod" validate:"oneof=dev prod test"`
	HTTPServer HTTPServer     `yaml:"http_server" validate:"required"`
	Postgres   PostgresConfig `yaml:"postgres" validate:"required"`
	Redis      RedisConfig    `yaml:"redis" validate:"required"`
}

func New() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read config from env: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return &cfg, nil
}
