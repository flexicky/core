package auth

import (
	"context"
	auth2 "core/internal/const/auth"
	grpcEnum "core/internal/const/grpc"
	authDto "core/internal/dto/auth"
	userDto "core/internal/dto/user"
	"core/internal/service/auth"
	"core/internal/service/user"
	"fmt"

	corev1 "github.com/flexicky/protos/gen/go/proto/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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

func (s *serverApi) extractClientInfo(ctx context.Context) (userAgent string, ipAddress string) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return "unknown", "unknown"
	}

	if userAgents := md.Get("user-agent"); len(userAgents) > 0 {
		userAgent = userAgents[0]
	} else {
		userAgent = "unknown"
	}

	if ips := md.Get("x-forwarded-for"); len(ips) > 0 {
		ipAddress = ips[0]
	} else if ips := md.Get("x-real-ip"); len(ips) > 0 {
		ipAddress = ips[0]
	} else {
		ipAddress = "unknown"
	}

	return userAgent, ipAddress
}

func (s *serverApi) Login(
	ctx context.Context,
	req *corev1.LoginRequest,
) (*corev1.LoginResponse, error) {

	params := authDto.Login{Email: req.Email, Password: req.Password}

	userAgent, ipAddress := s.extractClientInfo(ctx)

	ctx = context.WithValue(ctx, grpcEnum.UserAgent, userAgent)
	ctx = context.WithValue(ctx, grpcEnum.IPAddress, ipAddress)

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
