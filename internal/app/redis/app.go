package redisapp

import (
	"context"
	"fmt"
	"log/slog"

	redisAdapter "core/internal/redis"

	"github.com/redis/go-redis/v9"
)

type App struct {
	adapter *redisAdapter.RedisAdapter
	log     *slog.Logger
}

func New(log *slog.Logger, cfg redisAdapter.RedisConfig) (*App, error) {
	adapter, err := redisAdapter.NewRedisAdapter(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("init redis: %w", err)
	}

	return &App{adapter: adapter, log: log}, nil
}

func (a *App) Client() *redis.Client {
	return a.adapter.Client()
}

func (a *App) Stop(_ context.Context) error {
	if err := a.adapter.Stop(); err != nil {
		return fmt.Errorf("close redis: %w", err)
	}

	a.log.Info("redis connection closed")

	return nil
}
