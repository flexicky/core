package auth

import (
	"context"
	authAction "core/internal/action/auth"
	"core/internal/action/register"
	grpcEnum "core/internal/const/grpc"
	authDto "core/internal/dto/auth"
	userDto "core/internal/dto/user"
	"core/internal/utils"

	corev1 "github.com/flexicky/protos/core.core.v1"
	"google.golang.org/grpc"
)

type serverApi struct {
	corev1.UnimplementedAuthServer
	authAction     authAction.AuthAction
	registerAction register.RegisterAction
}

func RegisterServerAPI(gRPC *grpc.Server, authAction authAction.AuthAction, registerAction register.RegisterAction) {
	corev1.RegisterAuthServer(gRPC, &serverApi{

		authAction:     authAction,
		registerAction: registerAction,
	})
}

func (s *serverApi) Login(
	ctx context.Context,
	req *corev1.LoginRequest,
) (*corev1.LoginResponse, error) {
	params := authDto.Login{Email: req.Email, Password: req.Password}

	userAgent, ipAddress := utils.ExtractClientInfo(ctx)

	ctx = context.WithValue(ctx, grpcEnum.UserAgent, userAgent)
	ctx = context.WithValue(ctx, grpcEnum.IPAddress, ipAddress)

	accessToken, err := s.authAction.Run(ctx, params)
	if err != nil {
		errorString := err.Error()
		return &corev1.LoginResponse{
			Ok:      false,
			Message: &errorString,
		}, nil
	}

	return &corev1.LoginResponse{
		Ok:          true,
		AccessToken: &accessToken.Access_token,
	}, nil
}

func (s *serverApi) Register(
	ctx context.Context,
	req *corev1.RegisterRequest,
) (*corev1.RegisterResponse, error) {
	params := userDto.NewUser{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	return s.registerAction.Run(ctx, params), nil

}
