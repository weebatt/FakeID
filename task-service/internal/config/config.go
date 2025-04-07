package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPServer struct {
	Host string `yaml:"host" env:"HOST" env-default:"localhost" validate:"required"`
	Port string `yaml:"port" env:"PORT" env-default:"8080" validate:"required,numeric"`
}

type MongoConfig struct {
	URI          string `yaml:"uri" env:"MONGO_URI" env-default:"mongodb://localhost:27017" validate:"required,uri"`
	Database     string `yaml:"database" env:"MONGO_DATABASE" env-default:"myapp" validate:"required"`
	Username     string `yaml:"username" env:"MONGO_USERNAME" env-default:""`
	Password     string `yaml:"password" env:"MONGO_PASSWORD" env-default:""`
	Timeout      int    `yaml:"timeout" env:"MONGO_TIMEOUT" env-default:"5" validate:"gte=1"`
	PoolSize     uint64 `yaml:"pool_size" env:"MONGO_POOL_SIZE" env-default:"100" validate:"gte=1"`
	MaxConnIdle  int    `yaml:"max_conn_idle" env:"MONGO_MAX_CONN_IDLE" env-default:"10" validate:"gte=1"`
	MaxConnOpen  int    `yaml:"max_conn_open" env:"MONGO_MAX_CONN_OPEN" env-default:"50" validate:"gte=1"`
	ConnLifetime int    `yaml:"conn_lifetime" env:"MONGO_CONN_LIFETIME" env-default:"30" validate:"gte=1"`
}

type Config struct {
	Env        string      `yaml:"env" env:"ENV" env-default:"prod" validate:"oneof=dev prod test"`
	HTTPServer HTTPServer  `yaml:"http_server" validate:"required"`
	Mongo      MongoConfig `yaml:"mongo" validate:"required"`
}

func New() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		flag.StringVar(&configPath, "config", "./config/local.yaml", "path to config file")
		flag.Parse()
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return &cfg, nil
}
