package session

import (
	"context"
	postgreStorage "core/internal/storage"
	"time"
)

type Session struct {
	Id           int
	UserId       int
	CreatedAt    time.Time
	LastUsedAt   time.Time
	ExpiresAt    time.Time
	RefreshToken string
	RevokedAt    time.Time
	UserAgent    string
	IPAddress    string
}
type SessionCreate struct {
	UserId       int
	RefreshToken string
	ExpiryAt     time.Time
	UserAgent    string
	IpAddress    string
}

type sessionRepo struct {
	pool *postgreStorage.Storage
}

type SessionRepo interface {
	CreateSession(ctx context.Context, params SessionCreate) (*Session, error)
}

func NewSessionService(st *postgreStorage.Storage) SessionRepo {
	return &sessionRepo{pool: st}
}

func (r *sessionRepo) CreateSession(ctx context.Context, params SessionCreate) (*Session, error) {
	query := `
		INSERT INTO sessions 
    	(user_id, refresh_token, user_agent, ip_address, expires_at) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id 
	`
	session := &Session{
		RefreshToken: params.RefreshToken,
		UserId:       params.UserId,
		ExpiresAt:    params.ExpiryAt,
		UserAgent:    params.UserAgent,
		IPAddress:    params.IpAddress,
	}
	err := r.pool.Pool().QueryRow(ctx, query,
		session.UserId,
		session.RefreshToken,
		session.UserAgent,
		session.IPAddress,
		session.ExpiresAt,
	).Scan(&session.Id)
	if err != nil {
		return nil, err
	}

	return session, nil
}
