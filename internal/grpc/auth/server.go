package auth

import (
	"context"
	userDto "core/internal/dto/user"
	userRepo "core/internal/repository/user"
	"core/internal/service/auth"
	"core/internal/service/user"

	corev1 "github.com/flexicky/protos/gen/go/proto/core"
	"google.golang.org/grpc"
)

type serverApi struct {
	corev1.UnimplementedAuthServer
	userService user.UserSercive
	authService auth.AuthService
	userRepo    userRepo.UserRepository
}

func RegisterServerAPI(gRPC *grpc.Server, userService user.UserSercive) {
	corev1.RegisterAuthServer(gRPC, &serverApi{
		userService: userService,
	})
}

func (s *serverApi) Login(
	ctx context.Context,
	req *corev1.LoginRequest,
) (*corev1.LoginResponse, error) {

	return &corev1.LoginResponse{
		Token: req.GetEmail(),
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
