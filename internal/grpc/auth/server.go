package auth

import (
	"context"
	"core/internal/repository"
	"core/internal/service"

	corev1 "github.com/flexicky/protos/gen/go/proto/core"
	"google.golang.org/grpc"
)

type serverApi struct {
	corev1.UnimplementedAuthServer
	userService service.UserSercive
}

func RegisterServerAPI(gRPC *grpc.Server) {
	corev1.RegisterAuthServer(gRPC, &serverApi{})
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
	params := repository.UserCreateParams{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	user, err := s.userService.CreateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return &corev1.RegisterResponse{Id: user.Id}, nil
}
