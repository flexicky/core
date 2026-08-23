package postgresapp

import (
	"context"
	postgresStorage "core/internal/storage"
	"fmt"
	"log/slog"
	"time"
)

type App struct {
	log     *slog.Logger
	storage postgresStorage.Storage
	cfg     postgresStorage.DBConfig
}

func New(
	log *slog.Logger,
	cfg postgresStorage.DBConfig,
) (*App, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	storageInstance, err := postgresStorage.New(ctx, cfg, log)
	if err != nil {
		return nil, fmt.Errorf("does not init db pull: %w", err)
	}

	return &App{
		log:     log,
		storage: *storageInstance,
		cfg:     cfg,
	}, nil
}
