package auth

import (
	"context"
	authType "core/internal/const/auth"
	authDto "core/internal/dto/auth"
	authService "core/internal/service/auth"
	"fmt"
)

type authAction struct {
	authServ authService.AuthService
}

type AuthAction interface {
	Run(ctx context.Context, loginParams authDto.Login) (*authDto.LoginResult, error)
}

func NewAuthAction(authServ authService.AuthService) AuthAction {
	return &authAction{
		authServ: authServ,
	}
}

func (a *authAction) Run(ctx context.Context, loginParams authDto.Login) (*authDto.LoginResult, error) {
	result, err := a.authServ.Login(ctx, authType.EmailAuth, loginParams)
	if err != nil {
		return nil, fmt.Errorf("auth failed %w", err)
	}

	return result, nil
}
