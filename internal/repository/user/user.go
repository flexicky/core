package user

import (
	"context"
	"errors"
	"fmt"

	userDto "core/internal/dto/user"
	"core/internal/repository"

	"github.com/jackc/pgx/v5"
)

// Repository is the persistence port for users.
type Repository interface {
	Create(ctx context.Context, params userDto.NewUser) (*userDto.User, error)
	GetByID(ctx context.Context, id int64) (*userDto.User, error)
	GetByEmail(ctx context.Context, email string) (*userDto.User, error)
}

type repo struct {
	db repository.Querier
}

func New(db repository.Querier) Repository {
	return &repo{db: db}
}

// columns is the single source of truth for the column order used by scans.
// Explicit columns are preferred over SELECT * because a changed table layout
// silently breaks positional scans.
const columns = `id, name, email, pass, telegram_id, telegram_username, max_id, max_username, created_at`

func scan(row pgx.Row) (*userDto.User, error) {
	u := &userDto.User{}

	err := row.Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.Password,
		&u.TelegramID,
		&u.TelegramUsername,
		&u.MaxID,
		&u.MaxUsername,
		&u.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}

	return u, nil
}

func (r *repo) Create(ctx context.Context, params userDto.NewUser) (*userDto.User, error) {
	const query = `
		INSERT INTO users (name, email, pass, telegram_id, telegram_username, max_id, max_username)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING ` + columns

	created, err := scan(r.db.QueryRow(
		ctx,
		query,
		nullable(params.Name),
		nullable(params.Email),
		nullable(params.Password),
		nullable(params.TelegramID),
		nullable(params.TelegramUsername),
		nullable(params.MaxID),
		nullable(params.MaxUsername),
	))
	if err != nil {
		return nil, repository.Translate(err)
	}

	return created, nil
}

func (r *repo) GetByID(ctx context.Context, id int64) (*userDto.User, error) {
	const query = `SELECT ` + columns + ` FROM users WHERE id = $1`

	return scan(r.db.QueryRow(ctx, query, id))
}

func (r *repo) GetByEmail(ctx context.Context, email string) (*userDto.User, error) {
	const query = `SELECT ` + columns + ` FROM users WHERE email = $1`

	return scan(r.db.QueryRow(ctx, query, email))
}

// nullable stores an empty string as SQL NULL.
func nullable(value string) any {
	if value == "" {
		return nil
	}

	return value
}
