package redis

import (
	"context"
	"core/internal/redis"
	"fmt"
	"log/slog"
)

type App struct {
	adapter redis.RedisAdapter
	log     *slog.Logger
	cfg     redis.RedisConfig
}

func New(adapter redis.RedisAdapter, log *slog.Logger, cfg redis.RedisConfig) *App {
	return &App{
		adapter: adapter,
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
