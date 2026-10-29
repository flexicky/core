package postgresapp

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"core/internal/storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

const connectTimeout = 5 * time.Second

type App struct {
	log     *slog.Logger
	storage *storage.Storage
}

func New(log *slog.Logger, cfg storage.DBConfig) (*App, error) {
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	instance, err := storage.New(ctx, cfg, log)
	if err != nil {
		return nil, fmt.Errorf("init postgres: %w", err)
	}

	return &App{log: log, storage: instance}, nil
}

func (a *App) Pool() *pgxpool.Pool {
	return a.storage.Pool()
}

func (a *App) Stop(_ context.Context) error {
	a.storage.Close()

	return nil
}
