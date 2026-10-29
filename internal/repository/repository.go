// Package repository defines the database access boundary used by the
// concrete repositories (users, sessions) and their shared error values.
package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	// ErrNotFound is returned when a requested row does not exist.
	ErrNotFound = errors.New("repository: record not found")
	// ErrAlreadyExists is returned when a unique constraint is violated.
	ErrAlreadyExists = errors.New("repository: record already exists")
)

// Querier is the subset of *pgxpool.Pool used by the repositories.
// Depending on the interface (instead of the concrete pool) keeps the
// repositories easy to test with fakes.
type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// uniqueViolationCode is the PostgreSQL error code for a unique constraint violation.
const uniqueViolationCode = "23505"

// Translate converts driver level errors into repository errors.
func Translate(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
		return errors.Join(ErrAlreadyExists, err)
	}

	return err
}
