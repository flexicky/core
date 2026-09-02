package redis

import (
	"context"
	redisAdapter "core/internal/redis"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

type App struct {
	adapter redisAdapter.RedisAdapter
	log     *slog.Logger
	cfg     redisAdapter.RedisConfig
}

func New(log *slog.Logger, cfg redisAdapter.RedisConfig) *App {
	adapter, err := redisAdapter.NewRedisAdapter(cfg, log)
	if err != nil {
		panic("redis init is failed" + err.Error())
	}
	return &App{
		adapter: *adapter,
		log:     log,
		cfg:     cfg,
	}
}

func (a *App) Stop(ctx context.Context) error {
	done := make(chan struct{})

	go func() {
		if err := a.adapter.Stop(); err != nil {
			a.log.Error("redis close error", err)
		}
		close(done)
	}()

	select {
	case <-done:
		a.log.Info("redis closed successfully")
		return nil

	case <-ctx.Done():
		a.log.Warn("redis close timeout", "error", ctx.Err())
		return fmt.Errorf("redis close timeout: %w", ctx.Err())
	}
}

func (a *App) GetClient() *redis.Client {
	return a.adapter.GetRedisClient()
}
