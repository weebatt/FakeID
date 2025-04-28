package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"time"
)

type Config struct {
	Server struct {
		Port         string
		Host         string
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
	}

	Database struct {
		Host     string
		Port     string
		User     string
		Password string
		DBName   string
		SSLMode  string
	}

	JWT struct {
		Secret        string
		TokenExpiry   time.Duration
		RefreshExpiry time.Duration
	}

	Redis struct {
		Host     string
		Port     string
		Password string
		DB       int
	}

	Environment string
}

func Load() (*Config, error) {
	godotenv.Load() // Load .env if exists

	cfg := &Config{}

	// Server config
	cfg.Server.Port = getEnv("SERVER_PORT", "8080")
	cfg.Server.Host = getEnv("SERVER_HOST", "0.0.0.0")
	cfg.Server.ReadTimeout = time.Second * 15
	cfg.Server.WriteTimeout = time.Second * 15

	// Database config
	cfg.Database.Host = getEnv("PG_HOST", "localhost")
	cfg.Database.Port = getEnv("PG_PORT", "5432")
	cfg.Database.User = getEnv("PG_USER", "postgres")
	cfg.Database.Password = getEnv("PG_PASSWORD", "root")
	cfg.Database.DBName = getEnv("PG_DBNAME", "auth_service")
	cfg.Database.SSLMode = getEnv("PG_SSLMODE", "disable")

	// JWT config
	cfg.JWT.Secret = getEnv("JWT_SECRET", "your-secret-key")
	cfg.JWT.TokenExpiry = time.Hour * 24    // 24 hours
	cfg.JWT.RefreshExpiry = time.Hour * 168 // 7 days

	// Redis config
	cfg.Redis.Host = getEnv("REDIS_HOST", "localhost")
	cfg.Redis.Port = getEnv("REDIS_PORT", "6379")
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.DB = 0 // Redis DB index, can be configured via env if needed

	cfg.Environment = getEnv("ENV", "development")

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}
