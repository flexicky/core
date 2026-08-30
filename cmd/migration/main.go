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
		if err := createFile(createVal); err != nil {
			panic(err)
		}

		return
	}

	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	log.Info("starting migrations")

	wd, err := os.Getwd()
	if err != nil {
		log.Error("failed to get working directory", "error", err)
		os.Exit(1)
	}

	log.Info("working directory", "dir", wd)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Minute,
	)
	defer cancel()

	pool, err := postgresStorage.New(
		ctx,
		postgresStorage.DBConfig{
			Host:     cfg.DB.Host,
			User:     cfg.DB.User,
			Password: cfg.DB.Password,
			Name:     cfg.DB.Name,
			SSLMode:  cfg.DB.SslMode,
			Port:     cfg.DB.Port,
		},
		log,
	)

	if err != nil {
		log.Error(
			"failed to connect to postgres",
			"error",
			err,
		)

		os.Exit(1)
	}

	defer func() {
		log.Info("closing postgres connection")
		pool.Close()
		log.Info("postgres connection closed")
	}()

	dir := "./migrations"
	pattern := filepath.Join(dir, "*.sql")

	log.Info(
		"searching migrations",
		"dir",
		dir,
		"pattern",
		pattern,
	)

	files, err := filepath.Glob(pattern)
	if err != nil {
		log.Error(
			"failed to find migration files",
			"error",
			err,
		)

		os.Exit(1)
	}

	if len(files) == 0 {
		log.Error(
			"no migration files found",
			"dir",
			dir,
			"pattern",
			pattern,
		)

		os.Exit(1)
	}

	sort.Slice(files, func(i, j int) bool {
		name1 := filepath.Base(files[i])
		name2 := filepath.Base(files[j])

		parts1 := strings.SplitN(name1, "_", 2)
		parts2 := strings.SplitN(name2, "_", 2)

		if len(parts1) == 0 || len(parts2) == 0 {
			return name1 < name2
		}

		num1, err1 := strconv.Atoi(parts1[0])
		num2, err2 := strconv.Atoi(parts2[0])

		if err1 != nil || err2 != nil {
			return name1 < name2
		}

		return num1 < num2
	})

	log.Info(
		"migration files found",
		"count",
		len(files),
	)

	for _, file := range files {
		log.Info(
			"migration file",
			"file",
			file,
		)
	}

	migrations := make([]any, 0, len(files))

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			log.Error(
				"cannot read migration file",
				"file",
				file,
				"error",
				err,
			)

			os.Exit(1)
		}

		sql := strings.TrimSpace(string(data))

		if sql == "" {
			log.Warn(
				"migration file is empty",
				"file",
				file,
			)

			continue
		}

		migrationFile := file
		migrationSQL := sql

		log.Info(
			"migration loaded",
			"file",
			migrationFile,
			"bytes",
			len(data),
		)

		migrations = append(
			migrations,
			&migrator.Migration{
				Name: filepath.Base(migrationFile),

				Func: func(
					ctx context.Context,
					tx pgx.Tx,
				) error {
					log.Info(
						"executing migration",
						"file",
						migrationFile,
					)

					statements := strings.Split(
						migrationSQL,
						";",
					)

					for _, statement := range statements {
						stmt := strings.TrimSpace(statement)

						if stmt == "" {
							continue
						}

						log.Info(
							"executing SQL statement",
							"file",
							migrationFile,
							"sql",
							stmt,
						)

						if _, err := tx.Exec(ctx, stmt); err != nil {
							return fmt.Errorf(
								"executing statement in %s: %w",
								migrationFile,
								err,
							)
						}
					}

					log.Info(
						"migration executed",
						"file",
						migrationFile,
					)

					return nil
				},
			},
		)
	}

	if len(migrations) == 0 {
		log.Error(
			"no non-empty migrations available",
			"files_found",
			len(files),
		)

		os.Exit(1)
	}

	log.Info(
		"migrations loaded",
		"count",
		len(migrations),
	)

	m, err := migrator.New(
		migrator.Migrations(migrations...),
	)

	if err != nil {
		log.Error(
			"failed to create migrator",
			"error",
			err,
		)

		os.Exit(1)
	}

	log.Info(
		"applying migrations",
		"count",
		len(migrations),
	)

	if err := m.Migrate(ctx, pool.Pool()); err != nil {
		log.Error(
			"migration failed",
			"error",
			err,
		)

		os.Exit(1)
	}

	log.Info(
		"all migrations applied successfully",
	)
}

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func setupLogger(env string) *slog.Logger {
	switch env {
	case envLocal:
		return slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{
					Level: slog.LevelDebug,
				},
			),
		)

	case envDev:
		return slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{
					Level: slog.LevelDebug,
				},
			),
		)

	case envProd:
		return slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{
					Level: slog.LevelInfo,
				},
			),
		)

	default:
		return slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{
					Level: slog.LevelInfo,
				},
			),
		)
	}
}

func createFile(name string) error {
	const dir = "./migrations"

	name = strings.TrimSpace(name)

	if name == "" {
		return fmt.Errorf("migration name cannot be empty")
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf(
			"create migrations dir: %w",
			err,
		)
	}

	files, err := filepath.Glob(
		filepath.Join(dir, "*.sql"),
	)
	if err != nil {
		return fmt.Errorf(
			"glob migrations: %w",
			err,
		)
	}

	maxNum := 0

	for _, f := range files {
		base := filepath.Base(f)

		parts := strings.SplitN(
			base,
			"_",
			2,
		)

		if len(parts) == 0 {
			continue
		}

		n, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		if n > maxNum {
			maxNum = n
		}
	}

	next := maxNum + 1

	fileName := fmt.Sprintf(
		"%03d_migration_%s.sql",
		next,
		name,
	)

	fullPath := filepath.Join(
		dir,
		fileName,
	)

	if _, err := os.Stat(fullPath); err == nil {
		return fmt.Errorf(
			"migration %s already exists",
			fullPath,
		)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf(
			"check migration file: %w",
			err,
		)
	}

	if err := os.WriteFile(
		fullPath,
		[]byte(""),
		0o644,
	); err != nil {
		return fmt.Errorf(
			"write migration file: %w",
			err,
		)
	}

	fmt.Printf(
		"created migration: %s\n",
		fullPath,
	)

	return nil
}
