package auth

import (
	"context"
	authAction "core/internal/action/auth"
	"core/internal/action/me"
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
	MeAction       me.MeAction
}

func RegisterServerAPI(
	gRPC *grpc.Server,
	authAction authAction.AuthAction,
	registerAction register.RegisterAction,
	MeAction me.MeAction,
) {
	corev1.RegisterAuthServer(gRPC, &serverApi{

		authAction:     authAction,
		registerAction: registerAction,
		MeAction:       MeAction,
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

	return s.authAction.Run(ctx, params), nil
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

func (s *serverApi) Me(
	ctx context.Context,
	req *corev1.MeRequest,
) (*corev1.MeResponse, error) {
	userAgent, ipAddress := utils.ExtractClientInfo(ctx)
	ctx = context.WithValue(ctx, grpcEnum.UserAgent, userAgent)
	ctx = context.WithValue(ctx, grpcEnum.IPAddress, ipAddress)

	return s.MeAction.Run(ctx), nil
}
