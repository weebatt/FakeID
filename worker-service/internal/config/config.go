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

type KafkaConfig struct {
	Brokers    string `yaml:"brokers" env:"KAFKA_BROKERS" validate:"required"`
	Topic      string `yaml:"topic" env:"KAFKA_TOPIC" validate:"required"`
	MaxRetries int    `yaml:"max_retries" env:"KAFKA_MAX_RETRIES" env-default:"5" validate:"gte=1"`
	RetryDelay int    `yaml:"retry_delay" env:"KAFKA_RETRY_DELAY" env-default:"3" validate:"gte=1"`
	Timeout    int    `yaml:"timeout" env:"KAFKA_TIMEOUT" env-default:"5" validate:"gte=1"`
}

type Config struct {
	Env        string      `yaml:"env" env:"ENV" env-default:"prod" validate:"oneof=dev prod test"`
	HTTPServer HTTPServer  `yaml:"http_server" validate:"required"`
	Kafka      KafkaConfig `yaml:"kafka" validate:"required"`
}

func New() (*Config, error) {
	var cfg Config

	// Читаем конфигурацию из переменных окружения
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read config from env: %w", err)
	}

	// Валидация конфигурации
	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return &cfg, nil
}
