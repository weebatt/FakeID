package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *redis.Client
}

func NewRedis(addr, password string, db int) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Проверка подключения
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	return &Redis{Client: client}, nil
}

func (r *Redis) RedisClient() *redis.Client {
	return r.Client
}

func (r *Redis) Close() error {
	return r.Client.Close()
}
