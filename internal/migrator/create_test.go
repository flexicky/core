package migrator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateUsesNextFreeNumber(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	for _, name := range []string{"001_migration_users.sql", "002_migration_sessions.sql"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o600); err != nil {
			t.Fatalf("write fixture: %v", err)
		}
	}

	path, err := Create(dir, "products")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	want := filepath.Join(dir, "003_migration_products.sql")
	if path != want {
		t.Fatalf("Create() = %q, want %q", path, want)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("migration file was not created: %v", err)
	}
}

func TestCreateRejectsUnsafeNames(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"", "../escape", "UPPER", "with space", "drop;table"} {
		if _, err := Create(t.TempDir(), name); err == nil {
			t.Fatalf("Create(%q) error = nil, want error", name)
		}
	}
}

func TestMigrationFilesSorting(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	for _, name := range []string{"010_migration_ten.sql", "002_migration_two.sql", "001_migration_one.sql"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o600); err != nil {
			t.Fatalf("write fixture: %v", err)
		}
	}

	files, err := migrationFiles(dir)
	if err != nil {
		t.Fatalf("migrationFiles() error = %v", err)
	}

	want := []string{
		filepath.Join(dir, "001_migration_one.sql"),
		filepath.Join(dir, "002_migration_two.sql"),
		filepath.Join(dir, "010_migration_ten.sql"),
	}

	for i, file := range files {
		if file != want[i] {
			t.Fatalf("migrationFiles()[%d] = %q, want %q", i, file, want[i])
		}
	}
}
