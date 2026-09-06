package auth

import (
	"context"
	"core/internal/dto/auth"
	authService "core/internal/service/auth"
	userService "core/internal/service/user"
)

type authAction struct {
	authServ authService.AuthService
	userServ userService.UserSercive
}

type AuthService interface {
	Run(ctx context.Context, loginParams auth.Login)
}
