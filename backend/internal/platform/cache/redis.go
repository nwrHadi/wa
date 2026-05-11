package cache

import (
	"context"

	"wa-platform/backend/internal/config"

	"github.com/redis/go-redis/v9"
)

func NewRedis(ctx context.Context, cfg config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
