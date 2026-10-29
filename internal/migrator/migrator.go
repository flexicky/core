// Package migrator applies the SQL files in ./migrations.
package migrator

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	pgxmigrator "github.com/cybertec-postgresql/pgx-migrator"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DefaultDir is the directory the migration files live in.
const DefaultDir = "./migrations"

// Run applies every *.sql migration of dir in ascending numeric order.
func Run(ctx context.Context, log *slog.Logger, pool *pgxpool.Pool, dir string) error {
	files, err := migrationFiles(dir)
	if err != nil {
		return err
	}

	migrations := make([]any, 0, len(files))

	for _, path := range files {
		name := filepath.Base(path)

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		statements, err := SplitStatements(string(content))
		if err != nil {
			return fmt.Errorf("parse migration %s: %w", name, err)
		}

		if len(statements) == 0 {
			log.Warn("migration has no statements, skipping", slog.String("file", name))

			continue
		}

		migrations = append(migrations, &pgxmigrator.Migration{
			Name: name,
			Func: func(ctx context.Context, tx pgx.Tx) error {
				return apply(ctx, tx, name, statements)
			},
		})

		log.Debug("migration loaded", slog.String("file", name), slog.Int("statements", len(statements)))
	}

	if len(migrations) == 0 {
		return fmt.Errorf("no migrations found in %q", dir)
	}

	m, err := pgxmigrator.New(pgxmigrator.Migrations(migrations...))
	if err != nil {
		return fmt.Errorf("build migrator: %w", err)
	}

	log.Info("applying migrations", slog.Int("count", len(migrations)))

	if err := m.Migrate(ctx, pool); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	log.Info("migrations applied", slog.Int("count", len(migrations)))

	return nil
}

func apply(ctx context.Context, tx pgx.Tx, name string, statements []string) error {
	for _, statement := range statements {
		if _, err := tx.Exec(ctx, statement); err != nil {
			return fmt.Errorf("execute %s: %w", name, err)
		}
	}

	return nil
}

// migrationFiles returns the *.sql files of dir sorted by their numeric prefix.
func migrationFiles(dir string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return nil, fmt.Errorf("glob %q: %w", dir, err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no *.sql files found in %q", dir)
	}

	sort.SliceStable(files, func(i, j int) bool {
		left, right := filepath.Base(files[i]), filepath.Base(files[j])

		leftNum, leftErr := prefixNumber(left)
		rightNum, rightErr := prefixNumber(right)

		if leftErr != nil || rightErr != nil {
			return left < right
		}

		return leftNum < rightNum
	})

	return files, nil
}

func prefixNumber(name string) (int, error) {
	prefix, _, found := strings.Cut(name, "_")
	if !found {
		return 0, fmt.Errorf("migration %q has no numeric prefix", name)
	}

	return strconv.Atoi(prefix)
}
