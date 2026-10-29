package user

import (
	"context"
	"errors"
	"fmt"

	"core/internal/apperr"
	userDto "core/internal/dto/user"
	"core/internal/repository"
	"core/internal/repository/user"

	"golang.org/x/crypto/bcrypt"
)

// Service is the user domain API.
type Service interface {
	Create(ctx context.Context, params userDto.NewUser) (*userDto.User, error)
	GetByID(ctx context.Context, id int64) (*userDto.User, error)
	GetByEmail(ctx context.Context, email string) (*userDto.User, error)
	CheckPassword(password, hash string) bool
}

type service struct {
	repo user.Repository
}

func New(repo user.Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, params userDto.NewUser) (*userDto.User, error) {
	hash, err := hashPassword(params.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	params.Password = hash

	created, err := s.repo.Create(ctx, params)
	switch {
	case errors.Is(err, repository.ErrAlreadyExists):
		return nil, apperr.ErrUserAlreadyExists
	case err != nil:
		return nil, fmt.Errorf("create user: %w", err)
	}

	return created, nil
}

func (s *service) GetByID(ctx context.Context, id int64) (*userDto.User, error) {
	found, err := s.repo.GetByID(ctx, id)
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return nil, apperr.ErrUserNotFound
	case err != nil:
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return found, nil
}

func (s *service) GetByEmail(ctx context.Context, email string) (*userDto.User, error) {
	found, err := s.repo.GetByEmail(ctx, email)
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return nil, apperr.ErrUserNotFound
	case err != nil:
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return found, nil
}

// CheckPassword compares a plaintext password with a bcrypt hash.
func (s *service) CheckPassword(password, hash string) bool {
	if hash == "" {
		return false
	}

	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}
