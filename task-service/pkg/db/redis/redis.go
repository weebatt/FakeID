package redis

import (
	"context"
	"fmt"
	"task-service/internal/config"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Close() error
}

type Redis struct {
	Client *redis.Client
	logger *zap.SugaredLogger
}

func NewRedis(cfg config.RedisConfig, logger *zap.SugaredLogger) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		logger.Errorf("Failed to connect to Redis: %v", err)
		return nil, err
	}

	logger.Info("Successfully connected to Redis")
	return &Redis{Client: client, logger: logger}, nil
}

func (r *Redis) Close() error {
	r.logger.Info("Closing Redis connection")
	return r.Client.Close()
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}
