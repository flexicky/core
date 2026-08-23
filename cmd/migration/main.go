package main

import (
	"context"
	"core/internal/config"
	postgresStorage "core/internal/storage"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	migrator "github.com/cybertec-postgresql/pgx-migrator"
	"github.com/jackc/pgx/v5"
)

var migrationFS embed.FS

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting migrations")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := postgresStorage.New(ctx, postgresStorage.DBConfig{
		Host:     cfg.DB.Host,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		Name:     cfg.DB.Name,
		SSLMode:  cfg.DB.SslMode,
		Port:     cfg.DB.Port,
	}, log)

	if err != nil {
		log.Error("failed get pull postgresql : ", err)
	}

	defer pool.Close()

	var migrations []*migrator.Migration

	dir := "./migrations"
	pattern := filepath.Join(dir, "*.sql")

	files, err := filepath.Glob(pattern)
	if err != nil {
		log.Error("not found", err)
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			log.Warn("can not read file migration %s \n %w", file, err)
			continue
		}

		sql := string(data)

		if sql == "" {
			log.Warn("file migration is empty")
			continue
		}

		migrations = append(migrations, &migrator.Migration{
			Name: filepath.Base(file),
			Func: func(ctx context.Context, tx pgx.Tx) error {
				_, err := tx.Exec(ctx, sql)
				return err
			},
		})

		fmt.Println(sql)
	}
}

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
