package user

import (
	"context"
	"core/internal/repository/user"

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
	CreateUser(ctx context.Context, params user.UserCreateParams) (*user.User, error)
	GetUserById(ctx context.Context, id int) (*user.User, error)
	GetUserByEmail(ctx context.Context, email string) (*user.User, error)
	CheckPasswordHash(password, hash string) bool
}

type userService struct {
	repo user.UserRepository
}

func NewUserService(repo user.UserRepository) UserSercive {
	return &userService{repo: repo}
}

func (s *userService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *userService) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *userService) CreateUser(ctx context.Context, params user.UserCreateParams) (*user.User, error) {

	passwordHash, err := s.hashPassword(params.Password)
	if err != nil {
		return nil, err
	}

	params.Password = passwordHash

	return s.repo.Create(ctx, params)
}

func (s *userService) GetUserById(ctx context.Context, userId int) (*user.User, error) {
	user, err := s.repo.GetUserById(ctx, userId)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return user, nil
}
