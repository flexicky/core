package auth

import (
	"context"
	authAction "core/internal/action/auth"
	grpcEnum "core/internal/const/grpc"
	authDto "core/internal/dto/auth"
	userDto "core/internal/dto/user"
	"core/internal/service/auth"
	"core/internal/service/user"
	"core/internal/utils"
	"fmt"

	corev1 "github.com/flexicky/protos/core.core.v1"
	"google.golang.org/grpc"
)

type serverApi struct {
	corev1.UnimplementedAuthServer
	userService user.UserSercive
	authService auth.AuthService
	authAction  authAction.AuthAction
}

func RegisterServerAPI(gRPC *grpc.Server, userService user.UserSercive, authService auth.AuthService, authAction authAction.AuthAction) {
	corev1.RegisterAuthServer(gRPC, &serverApi{
		userService: userService,
		authService: authService,
		authAction:  authAction,
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

	user, err := s.userService.CreateUser(ctx, params)
	if err != nil {
		errorString := err.Error()
		return &corev1.RegisterResponse{
			Ok:      false,
			Message: &errorString,
		}, err
	}
	fmt.Print(user)
	return &corev1.RegisterResponse{
		Ok: true,
	}, nil
}
