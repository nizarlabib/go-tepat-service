package infrastructure

import (
	"context"
	"log"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(ctx context.Context, connection string) Redis {
	opts, err := redis.ParseURL(connection)
	if err != nil {
		log.Fatal("Failed to parsing redis connection: ", err)
	}
	client := redis.NewClient(opts)
	_, err = client.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Error when pinging to redis: ", err)
	}
	log.Println("Connection to Redis established successfully...")
	return Redis{
		client: client,
	}
}

func (infra *Redis) GetClient() *redis.Client {
	return infra.client
}

func (infra *Redis) Close() error {
	slog.Info("Closing redis connection...")
	return infra.client.Close()
}
