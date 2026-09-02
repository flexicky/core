package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	Prefix   string
}

type RedisAdapter struct {
	Client *redis.Client
	log    *slog.Logger
}

func connect(cfg RedisConfig, log *slog.Logger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   5,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	log.Info("Redis connected successfully", "host", cfg.Host, "port", cfg.Port)

	return client, nil
}

func NewRedisAdapter(cfg RedisConfig, log *slog.Logger) (*RedisAdapter, error) {
	redisConnection, err := connect(cfg, log)
	if err != nil {
		return nil, err
	}

	return &RedisAdapter{
		Client: redisConnection,
		log:    log,
	}, nil
}

func (r *RedisAdapter) Stop() error {
	if r.Client != nil {
		return r.Client.Close()
	}
	return nil
}
