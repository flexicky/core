package auth

type Login struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=6"`
}
