package auth

import (
	"context"
	auth2 "core/internal/const/auth"
	authDto "core/internal/dto/auth"
	userDto "core/internal/dto/user"
	"core/internal/service/auth"
	"core/internal/service/user"
	"fmt"

	corev1 "github.com/flexicky/protos/gen/go/proto/core"
	"google.golang.org/grpc"
)

type serverApi struct {
	corev1.UnimplementedAuthServer
	userService user.UserSercive
	authService auth.AuthService
}

func RegisterServerAPI(gRPC *grpc.Server, userService user.UserSercive, authService auth.AuthService) {
	corev1.RegisterAuthServer(gRPC, &serverApi{
		userService: userService,
		authService: authService,
	})
}

func (s *serverApi) Login(
	ctx context.Context,
	req *corev1.LoginRequest,
) (*corev1.LoginResponse, error) {

	params := authDto.Login{Email: req.Email, Password: req.Password}

	resutl, err := s.authService.Login(ctx, auth2.EmailAuth, params)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &corev1.LoginResponse{
		Token: resutl,
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
		return nil, err
	}

	return &corev1.RegisterResponse{Id: user.Id}, nil
}
