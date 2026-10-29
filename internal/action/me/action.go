package me

import (
	"context"
	"log/slog"
	"strconv"

	"core/internal/apperr"
	authService "core/internal/service/auth"
	"core/internal/utils"

	corev1 "github.com/flexicky/protos/core.core.v1"
)

type action struct {
	log         *slog.Logger
	authService authService.Service
}

// Action handles the Me RPC.
type Action interface {
	Run(ctx context.Context) *corev1.MeResponse
}

func New(log *slog.Logger, authService authService.Service) Action {
	return &action{log: log, authService: authService}
}

func (a *action) Run(ctx context.Context) *corev1.MeResponse {
	data, err := a.authService.CurrentSession(ctx)
	if err != nil {
		a.log.WarnContext(ctx, "get current session", slog.Any("error", err))

		return &corev1.MeResponse{Ok: false, Message: utils.Ptr(apperr.PublicMessage(err))}
	}

	return &corev1.MeResponse{
		Ok:     true,
		Name:   data.UserName,
		Email:  data.UserEmail,
		UserId: strconv.FormatInt(data.UserID, 10),
	}
}
