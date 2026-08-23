package main

import (
	"context"
	"core/internal/config"
	postgresStorage "core/internal/storage"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	migrator "github.com/cybertec-postgresql/pgx-migrator"
	"github.com/jackc/pgx/v5"
)

func main() {
	createPresent := false
	createVal := ""
	for i := 1; i < len(os.Args); i++ {
		a := os.Args[i]
		switch {
		case a == "--create":
			createPresent = true
			if i+1 < len(os.Args) {
				createVal = os.Args[i+1]
			}
		case strings.HasPrefix(a, "--create="):
			createPresent = true
			createVal = strings.TrimPrefix(a, "--create=")
		}
	}

	if createPresent && strings.TrimSpace(createVal) == "" {
		panic("flag --create must have a non-empty value, e.g. --create product")
	}
	if createPresent {
		createFile(createVal)
		return
	}

	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	log.Info("starting migrations")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
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
		log.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}

	defer pool.Close()

	var migrations []any

	dir := "./migrations"
	pattern := filepath.Join(dir, "*.sql")

	files, err := filepath.Glob(pattern)
	if err != nil {
		log.Warn("migrations not found")
		os.Exit(1)
	}

	sort.Slice(files, func(i, j int) bool {
		name1 := filepath.Base(files[i])
		name2 := filepath.Base(files[j])
		parts1 := strings.SplitN(name1, "_", 2)
		parts2 := strings.SplitN(name2, "_", 2)
		num1, _ := strconv.Atoi(parts1[0])
		num2, _ := strconv.Atoi(parts2[0])
		return num1 < num2
	})

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			log.Warn("can not read file migration", "file", file, "error", err)
			continue
		}

		sql := string(data)
		sqlCopy := sql

		if sql == "" {
			log.Warn("file migration is empty")
			continue
		}

		migrations = append(migrations, &migrator.Migration{
			Name: filepath.Base(file),
			Func: func(ctx context.Context, tx pgx.Tx) error {
				statements := strings.Split(sqlCopy, ";")

				for _, stmt := range statements {
					stmt := strings.TrimSpace(stmt)

					if stmt == "" {
						continue
					}

					if _, err := tx.Exec(ctx, stmt); err != nil {
						return fmt.Errorf("executing statement in %s: %w", filepath.Base(file), err)
					}
				}

				return nil
			},
		})
	}

	m, err := migrator.New(
		migrator.Migrations(migrations...),
	)
	if err != nil {
		log.Error("failed to create migrator", "error", err)
		os.Exit(1)
	}

	if err := m.Migrate(ctx, pool.Pool()); err != nil {
		log.Error("migration failed", "error", err)
		os.Exit(1)
	}

	log.Info("all migrations applied successfully")
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

func createFile(name string) error {
	const dir = "./migrations"

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create migrations dir: %w", err)
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}

	maxNum := 0

	for _, f := range files {
		base := filepath.Base(f)
		parts := strings.SplitN(base, "_", 2)
		n, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		if n > maxNum {
			maxNum = n
		}
	}

	next := maxNum + 1
	fileName := fmt.Sprintf("%03d_migration_%s.sql", next, name)
	fullPath := filepath.Join(dir, fileName)

	if _, err := os.Stat(fullPath); err == nil {
		return fmt.Errorf("migration %s already exists", fullPath)
	}

	template := fmt.Sprintf("-- Migration: %s\n-- Auto-generated\n\n", name)
	if err := os.WriteFile(fullPath, []byte(template), 0o644); err != nil {
		return fmt.Errorf("write migration file: %w", err)
	}

	return nil
}
