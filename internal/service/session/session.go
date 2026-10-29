package session

import (
	"context"
	"errors"
	"fmt"
	"time"

	"core/internal/apperr"
	sessionDto "core/internal/dto/session"
	"core/internal/repository"
	"core/internal/repository/session"
)

// Service is the session domain API.
type Service interface {
	Create(ctx context.Context, params sessionDto.NewSession) (*sessionDto.Session, error)
	GetByID(ctx context.Context, id int64) (*sessionDto.Session, error)
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}

type service struct {
	repo session.Repository
}

func New(repo session.Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, params sessionDto.NewSession) (*sessionDto.Session, error) {
	created, err := s.repo.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return created, nil
}

func (s *service) GetByID(ctx context.Context, id int64) (*sessionDto.Session, error) {
	found, err := s.repo.GetByID(ctx, id)
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return nil, apperr.ErrSessionNotFound
	case err != nil:
		return nil, fmt.Errorf("get session by id: %w", err)
	}

	return found, nil
}

// DeleteExpired removes sessions that are already expired.
func (s *service) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	deleted, err := s.repo.DeleteExpired(ctx, before)
	if err != nil {
		return 0, fmt.Errorf("delete expired sessions: %w", err)
	}

	return deleted, nil
}
