package user

import "time"

// NewUser is the input of the registration flow.
type NewUser struct {
	Name             string `validate:"omitempty,min=1,max=50"`
	Email            string `validate:"required,email,min=5,max=100"`
	Password         string `validate:"required,min=6,max=200"`
	TelegramID       string `validate:"omitempty,max=64"`
	TelegramUsername string `validate:"omitempty,max=64"`
	MaxID            string `validate:"omitempty,max=64"`
	MaxUsername      string `validate:"omitempty,max=64"`
}

// User is a stored user.
type User struct {
	ID               int64
	Name             *string
	Email            *string
	Password         *string
	TelegramID       *string
	TelegramUsername *string
	MaxID            *string
	MaxUsername      *string
	CreatedAt        time.Time
}
