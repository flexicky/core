package auth

import (
	"context"
	"log/slog"

	"core/internal/apperr"
	authType "core/internal/const/auth"
	authDto "core/internal/dto/auth"
	authService "core/internal/service/auth"
	"core/internal/utils"

	corev1 "github.com/flexicky/protos/core.core.v1"
)

type action struct {
	log      *slog.Logger
	authServ authService.Service
}

// Action handles the Login RPC.
type Action interface {
	Run(ctx context.Context, params authDto.Login) *corev1.LoginResponse
}

func New(log *slog.Logger, authServ authService.Service) Action {
	return &action{log: log, authServ: authServ}
}

func (a *action) Run(ctx context.Context, params authDto.Login) *corev1.LoginResponse {
	result, err := a.authServ.Login(ctx, authType.EmailAuth, params)
	if err != nil {
		a.log.WarnContext(ctx, "login failed", slog.String("email", params.Email), slog.Any("error", err))

		return &corev1.LoginResponse{Ok: false, Message: utils.Ptr(apperr.PublicMessage(err))}
	}

	return &corev1.LoginResponse{
		Ok:          true,
		AccessToken: utils.Ptr(result.AccessToken),
		ExpiresIn:   result.ExpiresAt.Unix(),
	}
}
