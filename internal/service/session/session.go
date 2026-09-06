package session

import (
	"context"
	sessionDto "core/internal/dto/session"
	"core/internal/repository/session"
	"core/internal/repository/user"
)

type sessionService struct {
	userRepo    user.UserRepository
	sessionRepo session.SessionRepo
}

type SessionService interface {
	CreateSession(ctx context.Context, params sessionDto.NewSession) (*sessionDto.Session, error)
}

func NewSessionService(userRepo user.UserRepository, sessionRepo session.SessionRepo) SessionService {
	return &sessionService{userRepo: userRepo, sessionRepo: sessionRepo}
}

func (s *sessionService) CreateSession(ctx context.Context, params sessionDto.NewSession) (*sessionDto.Session, error) {
	sessionData, err := s.sessionRepo.CreateSession(ctx, params)
	if err != nil {
		return &sessionDto.Session{}, err
	}

	return sessionData, nil
}

func (s *sessionService) GetSessionById(ctx context.Context, id int) (*sessionDto.Session, error) {
	session, err := s.sessionRepo.GetSessionById(ctx, id)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *sessionService) GetSessionByUserId(ctx context.Context, UserId int) (*sessionDto.Session, error) {
	session, err := s.sessionRepo.GetSessionByUserId(ctx, UserId)
	if err != nil {
		return nil, err
	}
	return session, nil
}
