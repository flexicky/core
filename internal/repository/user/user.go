package user

import (
	"context"
	userDto "core/internal/dto/user"
	postgreStorage "core/internal/storage"
	"fmt"
)

type UserRepository interface {
	Create(ctx context.Context, params userDto.NewUser) (*userDto.User, error)
	GetUserById(ctx context.Context, userId int) (*userDto.User, error)
	GetUserByEmail(ctx context.Context, email string) (*userDto.User, error)
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

func (r *userRepo) Create(ctx context.Context, params userDto.NewUser) (*userDto.User, error) {
	const query = `
		INSERT INTO users (name, email, pass, telegram_id, telegram_username, max_id, max_username)
		VALUES($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, pass
	`

	user := &userDto.User{
		Name:             &params.Name,
		Email:            &params.Email,
		TelegramId:       &params.TelegramId,
		TelegramUsername: &params.TelegramUsername,
		MaxId:            &params.MaxId,
		MaxUsername:      &params.MaxUsername,
	}

	err := r.pool.Pool().QueryRow(ctx, query,
		nilStringOrNil(*user.Name),
		nilStringOrNil(*user.Email),
		nilStringOrNil(params.Password),
		nilStringOrNil(*user.TelegramId),
		nilStringOrNil(*user.TelegramUsername),
		nilStringOrNil(*user.MaxId),
		nilStringOrNil(*user.MaxUsername),
	).Scan(&user.Id, &user.CreatedAt, &user.Password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (*userDto.User, error) {
	query := `
    SELECT *
    FROM users
    WHERE email = $1
`

	user := &userDto.User{}

	err := r.pool.Pool().QueryRow(ctx, query, email).Scan(&user.Id,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.TelegramId,
		&user.TelegramUsername,
		&user.MaxId,
		&user.MaxUsername,
		&user.CreatedAt,
	)
	if err != nil {

		fmt.Println("err seelcted ", err)
		return nil, err
	}

	return user, nil
}

func (r *userRepo) GetUserById(ctx context.Context, userId int) (*userDto.User, error) {
	query := `SELECT * FROM users WHERE id = $1`

	user := &userDto.User{}

	err := r.pool.Pool().QueryRow(ctx, query, userId).Scan(&user.Id, &user.Name, &user.Email, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}
