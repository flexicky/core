package auth

import "time"

// Login is the input of the login flow.
type Login struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=6,max=200"`
}

// LoginResult is a successfully issued access token.
type LoginResult struct {
	AccessToken string
	ExpiresAt   time.Time
}
