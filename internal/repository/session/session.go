package session

import (
	"context"
	"errors"
	"fmt"
	"time"

	sessionDto "core/internal/dto/session"
	"core/internal/repository"

	"github.com/jackc/pgx/v5"
)

// Repository is the persistence port for sessions.
type Repository interface {
	Create(ctx context.Context, params sessionDto.NewSession) (*sessionDto.Session, error)
	GetByID(ctx context.Context, id int64) (*sessionDto.Session, error)
	GetLatestByUserID(ctx context.Context, userID int64) (*sessionDto.Session, error)
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}

type repo struct {
	db repository.Querier
}

func New(db repository.Querier) Repository {
	return &repo{db: db}
}

const columns = `id, user_id, refresh_token, user_agent, ip_address, expires_at, revoked_at, created_at, last_used_at`

func scan(row pgx.Row) (*sessionDto.Session, error) {
	s := &sessionDto.Session{}

	err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.RefreshToken,
		&s.UserAgent,
		&s.IPAddress,
		&s.ExpiresAt,
		&s.RevokedAt,
		&s.CreatedAt,
		&s.LastUsedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan session: %w", err)
	}

	return s, nil
}

func (r *repo) Create(ctx context.Context, params sessionDto.NewSession) (*sessionDto.Session, error) {
	const query = `
		INSERT INTO sessions (user_id, refresh_token, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING ` + columns

	created, err := scan(r.db.QueryRow(
		ctx,
		query,
		params.UserID,
		params.RefreshToken,
		params.UserAgent,
		params.IPAddress,
		params.ExpiresAt,
	))
	if err != nil {
		return nil, repository.Translate(err)
	}

	return created, nil
}

func (r *repo) GetByID(ctx context.Context, id int64) (*sessionDto.Session, error) {
	const query = `SELECT ` + columns + ` FROM sessions WHERE id = $1`

	return scan(r.db.QueryRow(ctx, query, id))
}

func (r *repo) GetLatestByUserID(ctx context.Context, userID int64) (*sessionDto.Session, error) {
	const query = `SELECT ` + columns + `
		FROM sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1`

	return scan(r.db.QueryRow(ctx, query, userID))
}

// DeleteExpired removes every session that expired before the given moment and
// returns the number of deleted rows.
func (r *repo) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	const query = `DELETE FROM sessions WHERE expires_at < $1`

	tag, err := r.db.Exec(ctx, query, before)
	if err != nil {
		return 0, fmt.Errorf("delete expired sessions: %w", err)
	}

	return tag.RowsAffected(), nil
}
