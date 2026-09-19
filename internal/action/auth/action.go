package auth

import (
	"context"
	authType "core/internal/const/auth"
	authDto "core/internal/dto/auth"
	authService "core/internal/service/auth"
	"log/slog"

	corev1 "github.com/flexicky/protos/core.core.v1"
)

type authAction struct {
	log      *slog.Logger
	authServ authService.AuthService
}

type AuthAction interface {
	Run(ctx context.Context, loginParams authDto.Login) *corev1.LoginResponse
}

func NewAuthAction(log *slog.Logger, authServ authService.AuthService) AuthAction {
	return &authAction{
		log:      log,
		authServ: authServ,
	}
}

func (a *authAction) Run(ctx context.Context, loginParams authDto.Login) *corev1.LoginResponse {
	dto := &corev1.LoginResponse{}

	result, err := a.authServ.Login(ctx, authType.EmailAuth, loginParams)
	if err != nil {
		dto.Ok = false
		msg := err.Error()
		dto.Message = &msg
		go func() {
			a.log.Error("Auth failed", "error", err)
		}()
	}

	accessToken := result.Access_token
	exp := result.Expires_At

	dto.Ok = true
	dto.AccessToken = &accessToken
	dto.ExpiresIn = exp.Unix()

	return dto
}
