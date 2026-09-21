package me

import (
	"context"
	authService "core/internal/service/auth"
	"log/slog"
	"strconv"

	corev1 "github.com/flexicky/protos/core.core.v1"
)

type meAction struct {
	log         *slog.Logger
	authService authService.AuthService
}

type MeAction interface {
	Run(ctx context.Context) *corev1.MeResponse
}

func NewMeAction(log *slog.Logger, authService authService.AuthService) MeAction {
	return &meAction{
		log:         log,
		authService: authService,
	}
}

func (a *meAction) Run(ctx context.Context) *corev1.MeResponse {
	dto := &corev1.MeResponse{}

	data, err := a.authService.GetCurrentSession(ctx)
	if err != nil {
		dto.Ok = false
		mgs := err.Error()
		dto.Message = &mgs
		a.log.Warn(err.Error())
		return dto
	}

	dto.Ok = true
	dto.Name = data.UserName
	dto.UserId = strconv.Itoa(int(data.UserId))
	dto.Email = data.UserEmail

	return dto
}
