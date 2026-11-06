package infrastructure

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(ctx context.Context, connection string) (*Redis, error) {
	opts, err := redis.ParseURL(connection)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis connection: %w", err)
	}
	client := redis.NewClient(opts)
	if _, err = client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}
	slog.Info("Connection to Redis established successfully...")
	return &Redis{
		client: client,
	}, nil
}

func (infra *Redis) GetClient() *redis.Client {
	return infra.client
}

func (infra *Redis) Close() error {
	slog.Info("Closing redis connection...")
	if infra.client != nil {
		return infra.client.Close()
	}
	return nil
}
