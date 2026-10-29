// Package app assembles the application out of its infrastructure components.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	grpcapp "core/internal/app/grpc"
	postgresapp "core/internal/app/postgres"
	redisapp "core/internal/app/redis"
	redisAdapter "core/internal/redis"
	"core/internal/storage"
)

// Config is the application configuration passed to New.
type Config struct {
	GRPCPort               int
	AccessTokenTTL         time.Duration
	SessionTTL             time.Duration
	SessionCacheTTL        time.Duration
	SessionCleanupInterval time.Duration
	JWTSeed                string

	DB    storage.DBConfig
	Redis redisAdapter.RedisConfig
}

type App struct {
	grpc     *grpcapp.App
	postgres *postgresapp.App
	redis    *redisapp.App
}

// New builds the application. Every dependency is created eagerly so that a
// misconfigured service fails at start-up instead of on the first request.
func New(log *slog.Logger, cfg Config) (*App, error) {
	postgres, err := postgresapp.New(log, cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("postgres app: %w", err)
	}

	redisApp, err := redisapp.New(log, cfg.Redis)
	if err != nil {
		_ = postgres.Stop(context.Background())

		return nil, fmt.Errorf("redis app: %w", err)
	}

	grpcServer, err := grpcapp.New(log, grpcapp.Config{
		Port:                   cfg.GRPCPort,
		AccessTokenTTL:         cfg.AccessTokenTTL,
		SessionTTL:             cfg.SessionTTL,
		SessionCacheTTL:        cfg.SessionCacheTTL,
		SessionCleanupInterval: cfg.SessionCleanupInterval,
		RedisPrefix:            cfg.Redis.Prefix,
		JWTSeed:                cfg.JWTSeed,
	}, postgres.Pool(), redisApp.Client())
	if err != nil {
		_ = redisApp.Stop(context.Background())
		_ = postgres.Stop(context.Background())

		return nil, fmt.Errorf("grpc app: %w", err)
	}

	return &App{grpc: grpcServer, postgres: postgres, redis: redisApp}, nil
}

// Run serves requests until the process is stopped externally.
func (a *App) Run() error {
	return a.grpc.Run()
}

// Shutdown stops every component, attempting all of them even if one fails.
func (a *App) Shutdown(ctx context.Context) error {
	var errs []error

	if err := a.grpc.Stop(ctx); err != nil {
		errs = append(errs, fmt.Errorf("stop grpc: %w", err))
	}

	if err := a.postgres.Stop(ctx); err != nil {
		errs = append(errs, fmt.Errorf("stop postgres: %w", err))
	}

	if err := a.redis.Stop(ctx); err != nil {
		errs = append(errs, fmt.Errorf("stop redis: %w", err))
	}

	return errors.Join(errs...)
}
