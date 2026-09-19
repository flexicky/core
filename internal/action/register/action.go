package register

import (
	"context"
	authType "core/internal/const/auth"
	authDto "core/internal/dto/auth"
	userDto "core/internal/dto/user"
	"core/internal/service/auth"
	authServ "core/internal/service/auth"
	userServ "core/internal/service/user"
	"log/slog"

	corev1 "github.com/flexicky/protos/core.core.v1"
)

type registrerAction struct {
	log         *slog.Logger
	userService userServ.UserSercive
	authService auth.AuthService
}

type RegisterAction interface {
	Run(ctx context.Context, params userDto.NewUser) *corev1.RegisterResponse
}

func NewRegisterAction(log *slog.Logger, userService userServ.UserSercive, authService authServ.AuthService) RegisterAction {
	return &registrerAction{
		log:         log,
		userService: userService,
		authService: authService,
	}
}

func (a *registrerAction) Run(ctx context.Context, params userDto.NewUser) *corev1.RegisterResponse {
	dto := &corev1.RegisterResponse{}

	_, err := a.userService.CreateUser(ctx, params)

	if err != nil {
		msg := err.Error()
		dto.Ok = false
		dto.Message = &msg
		go func() {
			a.log.Error("Ошибка на строне Сервера", "error", err)
		}()
	}

	authResult, err := a.authService.Login(
		ctx,
		authType.EmailAuth,
		authDto.Login{
			Email:    params.Email,
			Password: params.Password,
		},
	)
	if err != nil {
		msg := err.Error()
		dto.Ok = false
		dto.Message = &msg
		go func() {
			a.log.Error("Auth failed", "error", err)
		}()
	}

	dto.Ok = true
	dto.Message = nil

	accessToken := authResult.Access_token
	exp := authResult.Expires_At

	dto.LoginData.AccessToken = &accessToken
	dto.LoginData.ExpiresIn = exp.Unix()

	return dto
}
