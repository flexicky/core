package session

import (
	"context"
	"core/internal/repository/user"
	"time"
)

type Session struct {
	Id           int
	RefreshToken string
	ExpiresAt    time.Time
	RevokedAT    time.Time
	LastUsedAt   time.Time
	UserAgent    string
	IpAddress    string
}
type sessionService struct {
	userRepo user.UserRepository
}

type SessionService interface {
	CreateSession(ctx context.Context, userId int) (string, error)
}

func NewSessionService(userRepo user.UserRepository) SessionService {
	return &sessionService{userRepo: userRepo}
}
