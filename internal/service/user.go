package service

import (
	"context"
	"core/internal/repository"

	"golang.org/x/crypto/bcrypt"
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

func (s *userService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *userService) checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *userService) CreateUser(ctx context.Context, params repository.UserCreateParams) (*repository.User, error) {

	passwordHash, err := s.hashPassword(params.Password)
	if err != nil {
		return nil, err
	}

	params.Password = passwordHash

	return s.repo.Create(ctx, params)
}

func (s *userService) GetUser(ctx context.Context, userId int) (*repository.User, error) {
	user, err := s.repo.GetUser(ctx, userId)
	if err != nil {
		return nil, err
	}

	return user, nil
}
