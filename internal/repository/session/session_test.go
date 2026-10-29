package session

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"core/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

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
	case *string:
		*ptr = value.(string)
	case *time.Time:
		*ptr = value.(time.Time)
	case **time.Time:
		*ptr = value.(*time.Time)
	default:
		return fmt.Errorf("assign: unsupported destination %T", target)
	}

	return nil
}

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

func TestGetByIDScansNullableTimestamps(t *testing.T) {
	t.Parallel()

	now := time.Now()
	db := &fakeQuerier{row: fakeRow{values: []any{
		int64(3),
		int64(9),
		"refresh-token",
		"grpc-go/1.83",
		"10.0.0.1",
		now.Add(time.Hour),
		(*time.Time)(nil),
		now,
		(*time.Time)(nil),
	}}}

	found, err := New(db).GetByID(context.Background(), 3)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if found.ID != 3 || found.UserID != 9 {
		t.Fatalf("scan result = %+v, want id=3 user_id=9", found)
	}

	if found.RevokedAt != nil || found.LastUsedAt != nil {
		t.Fatalf("nullable timestamps must stay nil, got %+v", found)
	}

	if found.CreatedAt.IsZero() {
		t.Fatal("created_at was not scanned")
	}

	if strings.Contains(db.query, "*") {
		t.Fatalf("query must not use SELECT *, got %q", db.query)
	}
}

func TestGetByIDMapsNoRows(t *testing.T) {
	t.Parallel()

	db := &fakeQuerier{row: fakeRow{err: pgx.ErrNoRows}}

	if _, err := New(db).GetByID(context.Background(), 1); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("GetByID() error = %v, want repository.ErrNotFound", err)
	}
}

func TestGetLatestByUserOrdersByCreatedAt(t *testing.T) {
	t.Parallel()

	db := &fakeQuerier{row: fakeRow{err: pgx.ErrNoRows}}

	_, _ = New(db).GetLatestByUserID(context.Background(), 5)

	normalized := strings.Join(strings.Fields(db.query), " ")
	if !strings.Contains(normalized, "ORDER BY created_at DESC LIMIT 1") {
		t.Fatalf("query = %q, want deterministic ordering", normalized)
	}
}
