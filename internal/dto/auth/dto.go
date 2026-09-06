package auth

import "time"

type Login struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=6"`
}

type LoginResult struct {
	Access_token string
	Expires_At   time.Time
}
