package register

import (
	"context"
	"log/slog"

	"core/internal/apperr"
	authType "core/internal/const/auth"
	authDto "core/internal/dto/auth"
	userDto "core/internal/dto/user"
	authService "core/internal/service/auth"
	userService "core/internal/service/user"
	"core/internal/utils"

	corev1 "github.com/flexicky/protos/core.core.v1"
)

type action struct {
	log         *slog.Logger
	userService userService.Service
	authService authService.Service
}

// Action handles the Register RPC.
type Action interface {
	Run(ctx context.Context, params userDto.NewUser) *corev1.RegisterResponse
}

func New(log *slog.Logger, userService userService.Service, authService authService.Service) Action {
	return &action{
		log:         log,
		userService: userService,
		authService: authService,
	}
}

func (a *action) Run(ctx context.Context, params userDto.NewUser) *corev1.RegisterResponse {
	if _, err := a.userService.Create(ctx, params); err != nil {
		a.log.ErrorContext(ctx, "register user", slog.String("email", params.Email), slog.Any("error", err))

		return &corev1.RegisterResponse{Ok: false, Message: utils.Ptr(apperr.PublicMessage(err))}
	}

	result, err := a.authService.Login(ctx, authType.EmailAuth, authDto.Login{
		Email:    params.Email,
		Password: params.Password,
	})
	if err != nil {
		a.log.ErrorContext(ctx, "login after register", slog.String("email", params.Email), slog.Any("error", err))

		return &corev1.RegisterResponse{Ok: false, Message: utils.Ptr(apperr.PublicMessage(err))}
	}

	return &corev1.RegisterResponse{
		Ok: true,
		LoginData: &corev1.LoginResponse{
			Ok:          true,
			AccessToken: utils.Ptr(result.AccessToken),
			ExpiresIn:   result.ExpiresAt.Unix(),
		},
	}
}
