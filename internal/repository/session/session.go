package session

import (
	"context"
	"core/internal/dto/session"
	postgreStorage "core/internal/storage"
)

type sessionRepo struct {
	pool *postgreStorage.Storage
}

type SessionRepo interface {
	CreateSession(ctx context.Context, params session.NewSession) (*session.Session, error)
}

func NewSessionRepo(st *postgreStorage.Storage) SessionRepo {
	return &sessionRepo{pool: st}
}

func (r *sessionRepo) CreateSession(ctx context.Context, params session.NewSession) (*session.Session, error) {
	query := `
		INSERT INTO sessions 
    	(user_id, refresh_token, user_agent, ip_address, expires_at) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id 
	`
	sessionData := &session.Session{
		RefreshToken: params.RefreshToken,
		UserId:       params.UserID,
		ExpiresAt:    params.ExpiresAt,
		UserAgent:    params.UserAgent,
		IPAddress:    params.IpAddress,
	}
	err := r.pool.Pool().QueryRow(ctx, query,
		sessionData.UserId,
		sessionData.RefreshToken,
		sessionData.UserAgent,
		sessionData.IPAddress,
		sessionData.ExpiresAt,
	).Scan(&sessionData.Id)
	if err != nil {
		return nil, err
	}

	return sessionData, nil
}
