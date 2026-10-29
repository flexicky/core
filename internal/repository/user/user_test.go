package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	userDto "core/internal/dto/user"
	"core/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// fakeRow is a minimal pgx.Row.
type fakeRow struct {
	values []any
	err    error
}

func (r fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}

	if len(dest) != len(r.values) {
		return fmt.Errorf("scan: got %d destinations, want %d", len(dest), len(r.values))
	}

	for i, target := range dest {
		if err := assign(target, r.values[i]); err != nil {
			return err
		}
	}

	return nil
}

func assign(target, value any) error {
	switch ptr := target.(type) {
	case *int64:
		*ptr = value.(int64)
	case **string:
		*ptr = value.(*string)
	case *time.Time:
		*ptr = value.(time.Time)
	default:
		return fmt.Errorf("assign: unsupported destination %T", target)
	}

	return nil
}

// fakeQuerier captures queries and returns canned rows.
type fakeQuerier struct {
	row   pgx.Row
	query string
	args  []any
}

func (f *fakeQuerier) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	f.query = sql
	f.args = args

	return f.row
}

func (f *fakeQuerier) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeQuerier) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("not implemented")
}

func ptr(value string) *string { return &value }

func TestGetByEmailUsesExplicitColumns(t *testing.T) {
	t.Parallel()

	db := &fakeQuerier{row: fakeRow{err: pgx.ErrNoRows}}

	if _, err := New(db).GetByEmail(context.Background(), "user@example.com"); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("GetByEmail() error = %v, want repository.ErrNotFound", err)
	}

	if strings.Contains(db.query, "*") {
		t.Fatalf("query must not use SELECT *, got %q", db.query)
	}

	for _, column := range []string{"id", "name", "email", "pass", "created_at"} {
		if !strings.Contains(db.query, column) {
			t.Fatalf("query %q is missing column %q", db.query, column)
		}
	}
}

func TestCreateScansReturnedRow(t *testing.T) {
	t.Parallel()

	createdAt := time.Now()
	db := &fakeQuerier{row: fakeRow{values: []any{
		int64(7),
		ptr("Test"),
		ptr("user@example.com"),
		ptr("hash"),
		(*string)(nil),
		(*string)(nil),
		(*string)(nil),
		(*string)(nil),
		createdAt,
	}}}

	created, err := New(db).Create(context.Background(), userDto.NewUser{
		Name:     "Test",
		Email:    "user@example.com",
		Password: "hash",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created.ID != 7 {
		t.Fatalf("ID = %d, want 7", created.ID)
	}

	if created.Email == nil || *created.Email != "user@example.com" {
		t.Fatalf("Email = %v, want user@example.com", created.Email)
	}

	if !created.CreatedAt.Equal(createdAt) {
		t.Fatalf("CreatedAt = %v, want %v", created.CreatedAt, createdAt)
	}
}

func TestCreateMapsUniqueViolation(t *testing.T) {
	t.Parallel()

	db := &fakeQuerier{row: fakeRow{err: &pgconn.PgError{Code: "23505", Message: "duplicate key"}}}

	_, err := New(db).Create(context.Background(), userDto.NewUser{Email: "dup@example.com"})
	if !errors.Is(err, repository.ErrAlreadyExists) {
		t.Fatalf("Create() error = %v, want repository.ErrAlreadyExists", err)
	}
}

func TestEmptyStringsAreStoredAsNull(t *testing.T) {
	t.Parallel()

	if nullable("") != nil {
		t.Fatal("nullable(\"\") must be nil")
	}

	if nullable("value") != "value" {
		t.Fatal("nullable(\"value\") must keep the value")
	}
}
