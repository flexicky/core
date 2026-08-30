package user

import "time"

type NewUser struct {
	Name             string `validate:"omitempty,min=1,max=50"`
	Email            string `validate:"email,omitempty,min=5,max=100"`
	TelegramId       string
	TelegramUsername string
	MaxId            string
	MaxUsername      string
	Password         string `validate:"omitempty,min=6,max=200"`
}

type User struct {
	Id               int64
	Name             *string
	Email            *string
	TelegramId       *string
	TelegramUsername *string
	MaxId            *string
	MaxUsername      *string
	CreatedAt        time.Time
	Password         *string
}
