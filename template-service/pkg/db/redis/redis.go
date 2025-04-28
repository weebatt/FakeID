package redis

import (
	"fmt"
	"template-service/internal/config"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type Redis struct {
	Client *redis.Client
	logger *zap.SugaredLogger
	cb     *gobreaker.CircuitBreaker
}

func NewRedis(cfg config.RedisConfig, logger *zap.SugaredLogger) (*Redis, error) {
	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		client := redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		})

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Timeout)*time.Second)
		defer cancel()

		_, err := client.Ping(ctx).Result()
		if err != nil {
			client.Close()
			logger.Warnf("Failed to connect to Redis on attempt %d: %v", attempt, err)
			if attempt == cfg.MaxRetries {
				return nil, fmt.Errorf("failed to connect to Redis after %d attempts: %w", cfg.MaxRetries, err)
			}
			time.Sleep(time.Duration(cfg.RetryDelay) * time.Second)
			continue
		}

		logger.Infof("Connected to Redis on attempt %d", attempt)
		cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:        "redis",
			MaxRequests: 1,
			Interval:    30 * time.Second,
			Timeout:     10 * time.Second,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures >= 3
			},
			OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
				logger.Infof("Redis circuit breaker state changed for %s: %s -> %s", name, from.String(), to.String())
			},
		})

		return &Redis{
			Client: client,
			logger: logger,
			cb:     cb,
		}, nil
	}

	return nil, fmt.Errorf("failed to connect to Redis after %d attempts", cfg.MaxRetries)
}

func (r *Redis) Close() {
	if err := r.Client.Close(); err != nil {
		r.logger.Warnf("Failed to close Redis client: %v", err)
	}
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	result, err := r.cb.Execute(func() (interface{}, error) {
		return r.Client.Get(ctx, key).Result()
	})
	if err != nil {
		r.logger.Errorf("Circuit Breaker rejected Get: %v", err)
		return "", err
	}

	return result.(string), nil
}

func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	_, err := r.cb.Execute(func() (interface{}, error) {
		return nil, r.Client.Set(ctx, key, value, expiration).Err()
	})
	if err != nil {
		r.logger.Errorf("Circuit Breaker rejected Set: %v", err)
	}
	return err
}
