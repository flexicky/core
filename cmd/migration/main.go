package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"core/internal/config"
	"core/internal/logger"
	"core/internal/migrator"
	"core/internal/storage"
)

const migrateTimeout = 5 * time.Minute

func main() {
	if err := run(); err != nil {
		slog.Error("migration failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", os.Getenv("CONFIG_PATH"), "path to the configuration file")
	create := flag.String("create", "", "create a new empty migration with the given name")
	dir := flag.String("dir", migrator.DefaultDir, "directory with migration files")
	flag.Parse()

	if *create != "" {
		path, err := migrator.Create(*dir, *create)
		if err != nil {
			return err
		}

		fmt.Printf("created migration: %s\n", path)

		return nil
	}

	cfg := config.MustLoad(*configPath)

	log := logger.New(cfg.Env).With(slog.String("service", "migration"))

	ctx, cancel := context.WithTimeout(context.Background(), migrateTimeout)
	defer cancel()

	pool, err := storage.New(ctx, storage.DBConfig{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		Name:     cfg.DB.Name,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		SSLMode:  cfg.DB.SSLMode,
	}, log)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer pool.Close()

	return migrator.Run(ctx, log, pool.Pool(), *dir)
}
