package health

import (
	"context"
	// "errors"
	"fmt"
	// "log/slog"
	"time"

	// "github.com/minio/minio-go/v7"
	// "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	ResourcePostgreSQL string = "postgresql"
	ResourceRabbitMQ   string = "rabbitmq"
	ResourceRedis      string = "redis"
	ResourceMinio      string = "minio"
)

func PostgreSQLHealthConfig(db *gorm.DB) Config {
	return Config{
		Name:      ResourcePostgreSQL,
		Timeout:   5 * time.Second,
		SkipOnErr: true,
		Check: func(ctx context.Context) error {
			err := db.WithContext(ctx).Exec("SELECT VERSION();").Error
			if err != nil {
				return fmt.Errorf("postgresql health check failed: %w", err)
			}
			return nil
		},
	}
}

// func RabbitMQHealthConfig(rmq *amqp091.Connection) Config {
// 	return Config{
// 		Name:      ResourceRabbitMQ,
// 		Timeout:   5 * time.Second,
// 		SkipOnErr: true,
// 		Check: func(ctx context.Context) error {
// 			slog.Debug("performing rabbitmq health check")
// 			if rmq.IsClosed() {
// 				return errors.New("rabbitmq health check connection is closed")
// 			}
// 			return nil
// 		},
// 	}
// }

func RedisHealthConfig(redisClient *redis.Client) Config {
	return Config{
		Name:      ResourceRedis,
		Timeout:   5 * time.Second,
		SkipOnErr: true,
		Check: func(ctx context.Context) error {
			pong, err := redisClient.Ping(ctx).Result()
			if err != nil {
				return fmt.Errorf("redis health check ping failed: %w", err)
			}
			if pong != "PONG" {
				return fmt.Errorf("redis health unexpected redis ping response: %s", pong)
			}
			return nil
		},
	}
}

// func MinioHealthConfig(minioClient *minio.Client, bucketName string) Config {
// 	return Config{
// 		Name:      ResourceMinio,
// 		Timeout:   5 * time.Second,
// 		SkipOnErr: true,
// 		Check: func(ctx context.Context) error {
// 			if bucketName == "" {
// 				_, err := minioClient.ListBuckets(ctx)
// 				if err != nil {
// 					return fmt.Errorf("minio health check list buckets failed: %w", err)
// 				}
// 				return nil
// 			}
// 			exists, err := minioClient.BucketExists(ctx, bucketName)
// 			if err != nil {
// 				return fmt.Errorf("minio health check failed: %w", err)
// 			}
// 			if !exists {
// 				return fmt.Errorf("minio health check bucket %s does not exist", bucketName)
// 			}
// 			return nil
// 		},
// 	}
// }
