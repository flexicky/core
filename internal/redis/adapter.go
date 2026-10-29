package redis

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	connectTimeout = 5 * time.Second
	poolSize       = 10
	minIdleConns   = 5
	maxRetries     = 5
)

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	Prefix   string
}

type RedisAdapter struct {
	client *redis.Client
	log    *slog.Logger
}

// NewRedisAdapter connects to Redis and verifies the connection with PING.
func NewRedisAdapter(cfg RedisConfig, log *slog.Logger) (*RedisAdapter, error) {
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     cfg.Password,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
		MaxRetries:   maxRetries,
	})

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()

		return nil, fmt.Errorf("ping redis at %s: %w", addr, err)
	}

	log.Info("redis connected", slog.String("addr", addr))

	return &RedisAdapter{client: client, log: log}, nil
}

func (a *RedisAdapter) Client() *redis.Client {
	return a.client
}

func (a *RedisAdapter) Stop() error {
	if a.client == nil {
		return nil
	}

	return a.client.Close()
}
