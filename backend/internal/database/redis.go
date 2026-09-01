package database

import (
	"context"
	"fmt"

	"github.com/community-platform/backend/internal/config"
	"github.com/redis/go-redis/v9"
)

// NewRedis creates and verifies a Redis client connection.
func NewRedis(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       0,
	})

	// Verify connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connecting to redis: %w", err)
	}

	return client, nil
}

// MustNewRedis is like NewRedis but panics on error. Use during startup.
func MustNewRedis(cfg config.RedisConfig) *redis.Client {
	client, err := NewRedis(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to redis: %v", err))
	}
	return client
}
