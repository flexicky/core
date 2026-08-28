package user

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
	Password         string
}

type UserCreateParams struct {
	Name             string `validate:"omitempty,min=1,max=50"`
	Email            string `validate:"email, omitempty,min=5,max=100"`
	TelegramId       string
	TelegramUsername string
	MaxId            string
	MaxUsername      string
	Password         string `validate:"omitempty,min=6,max=200"`
}

type UserRepository interface {
	Create(ctx context.Context, params UserCreateParams) (*User, error)
	GetUserById(ctx context.Context, userId int) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
}

type userRepo struct {
	pool *postgreStorage.Storage
}

func NewUserRepo(st *postgreStorage.Storage) UserRepository {
	return &userRepo{pool: st}
}

func nilStringOrNil(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (r *userRepo) Create(ctx context.Context, params UserCreateParams) (*User, error) {
	const query = `
		INSERT INTO users (name, email, pass, telegram_id, telegram_username, max_id, max_username)
		VALUES($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, pass
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
		nilStringOrNil(user.Name),
		nilStringOrNil(user.Email),
		nilStringOrNil(params.Password),
		nilStringOrNil(user.TelegramId),
		nilStringOrNil(user.TelegramUsername),
		nilStringOrNil(user.MaxId),
		nilStringOrNil(user.MaxUsername),
	).Scan(&user.Id, &user.CreatedAt, &user.Password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT * FROM users WHERE email = $1`

	user := &User{}

	err := r.pool.Pool().QueryRow(ctx, query, email).Scan(&user.Id, &user.Name, &user.Email, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepo) GetUserById(ctx context.Context, userId int) (*User, error) {
	query := `SELECT * FROM users WHERE id = $1`

	user := &User{}

	err := r.pool.Pool().QueryRow(ctx, query, userId).Scan(&user.Id, &user.Name, &user.Email, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}
