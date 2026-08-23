package repository

import (
	"context"
	postgreStorage "core/internal/storage"
	"time"
)

type User struct {
	Id               int64
	Name             string
	Email            string
	TelegramId       string
	TelegramUsername string
	MaxId            string
	MaxUsername      string
	CreatedAt        time.Time
}

type UserCreateParams struct {
	Name             string
	Email            string
	TelegramId       string
	TelegramUsername string
	MaxId            string
	MaxUsername      string
	Password         string
}

type UserRepository interface {
	Create(ctx context.Context, params UserCreateParams) (*User, error)
}

type userRepo struct {
	pool *postgreStorage.Storage
}

func New(st *postgreStorage.Storage) UserRepository {
	return &userRepo{pool: st}
}

func (r *userRepo) Create(ctx context.Context, params UserCreateParams) (*User, error) {
	const query = `
		INSERT INTO users (name, email, password, telegram_id, telegram_username, max_id, max_username)
		VALUES($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	user := &User{
		Name:             params.Name,
		Email:            params.Email,
		TelegramId:       params.TelegramId,
		TelegramUsername: params.TelegramUsername,
		MaxId:            params.MaxId,
		MaxUsername:      params.MaxUsername,
	}

	err := r.pool.Pool().QueryRow(ctx, query,
		user.Name,
		user.Email,
		params.Password,
		user.TelegramId,
		user.TelegramUsername,
		user.MaxId,
		user.MaxUsername,
	).Scan(&user.Id, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}
