package cache

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var log = slog.Default().With("service", "cache")

func RevalidateKeys(ctx context.Context, redis *redis.Client, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	_, err := redis.Del(ctx, keys...).Result()

	log.Info("cache revalidated", slog.String("keys", strings.Join(keys, ", ")))

	return err
}

func WithCache[T comparable](ctx context.Context, redis *redis.Client, key string, fetchFunc func() (T, error)) (T, error) {
	return WithCacheExpiry(ctx, redis, key, fetchFunc, 0)
}

func WithCacheExpiry[T comparable](ctx context.Context, redis *redis.Client, key string, fetchFunc func() (T, error), expirySeconds int) (T, error) {
	cachedData, err := FromCache[T](ctx, redis, key)
	if err == nil {
		return cachedData, nil
	}

	var zero T

	log.Info("cache miss", slog.String("key", key))

	data, err := fetchFunc()
	if err != nil {
		return zero, err
	}

	err = SaveToCache(ctx, redis, key, data, expirySeconds)
	if err != nil {
		return zero, err
	}

	return data, nil
}

func SaveToCache[T comparable](ctx context.Context, redis *redis.Client, key string, data T, expirySeconds int) error {
	marshaledData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = redis.Set(ctx, key, marshaledData, time.Duration(expirySeconds)*time.Second).Err()
	if err != nil {
		return err
	}

	return nil
}

func FromCache[T comparable](ctx context.Context, redis *redis.Client, key string) (T, error) {
	var zero T

	cachedData, err := redis.Get(ctx, key).Result()
	if err != nil {
		return zero, err
	}

	var result T
	err = json.Unmarshal([]byte(cachedData), &result)
	if err != nil {
		return zero, err
	}

	log.Info("cache hit", slog.String("key", key))

	return result, nil
}
