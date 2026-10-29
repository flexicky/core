package migrator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var migrationNamePattern = regexp.MustCompile(`^[a-z0-9_]+$`)

// Create makes an empty migration file with the next free numeric prefix, e.g.
// 003_migration_add_products.sql.
func Create(dir, name string) (string, error) {
	if !migrationNamePattern.MatchString(name) {
		return "", fmt.Errorf("migration name %q must match %s", name, migrationNamePattern)
	}

	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", fmt.Errorf("create migrations dir: %w", err)
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return "", fmt.Errorf("glob migrations: %w", err)
	}

	next := 1

	for _, file := range files {
		number, err := prefixNumber(filepath.Base(file))
		if err != nil {
			continue
		}

		if number >= next {
			next = number + 1
		}
	}

	path := filepath.Join(dir, fmt.Sprintf("%03d_migration_%s.sql", next, name))

	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("migration %q already exists", path)
	}

	if err := os.WriteFile(path, nil, 0o600); err != nil {
		return "", fmt.Errorf("write migration: %w", err)
	}

	return path, nil
}
