package service

import (
	"context"
	"core/internal/repository"
	"errors"
)

type UserCreateParams struct {
	Email            string
	Password         string
	Name             string
	TelegramId       string
	TelegramUsername string
	MaxId            string
	MaxUsername      string
}

type UserSercive interface {
	CreateUser(ctx context.Context, params repository.UserCreateParams) (*repository.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserSercive {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, params repository.UserCreateParams) (*repository.User, error) {
	//TODO вынести в отдельный валидатор
	if params.Email == "" || params.Password == "" {
		return nil, errors.New("email is password are required fields")
	}

	return s.repo.Create(ctx, params)
}
