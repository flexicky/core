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
	GetSessionById(ctx context.Context, id int) (*session.Session, error)
	GetSessionByUserId(ctx context.Context, userId int) (*session.Session, error)
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

func (r *sessionRepo) GetSessionById(ctx context.Context, id int) (*session.Session, error) {
	query := `SELECT * FROM sessions where id = $1`

	session := &session.Session{}

	err := r.pool.Pool().QueryRow(ctx, query, id).Scan(&session.Id, &session.UserId, &session.CreatedAt, &session.LastUsedAt, &session.ExpiresAt, &session.RefreshToken, &session.RevokedAt, &session.UserAgent, &session.IPAddress)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (r *sessionRepo) GetSessionByUserId(ctx context.Context, userId int) (*session.Session, error) {
	query := `SELECT * FROM sessions where user_id = $1`

	session := &session.Session{}

	err := r.pool.Pool().QueryRow(ctx, query, userId).Scan(&session.Id, &session.UserId, &session.CreatedAt, &session.LastUsedAt, &session.ExpiresAt, &session.RefreshToken, &session.RevokedAt, &session.UserAgent, &session.IPAddress)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (r *sessionRepo) GetSessionByExpiresAt(ctx context.Context, exp int64) ([]int64, error) {
	query := `SELECT id FROM sessions WHERE expires_at > $1`

	rows, err := r.pool.Pool().Query(ctx, query, exp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64

	for rows.Next() {
		var id int64

		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func (r *sessionRepo) RevokeSessionByUserId(ctx context.Context, userId int) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err := r.pool.Pool().Exec(ctx, query, userId)
	if err != nil {
		return err
	}
	return nil
}
