package auth

import (
	"context"

	authAction "core/internal/action/auth"
	meAction "core/internal/action/me"
	registerAction "core/internal/action/register"
	"core/internal/app/validator"
	"core/internal/apperr"
	"core/internal/ctxkeys"
	authDto "core/internal/dto/auth"
	userDto "core/internal/dto/user"
	"core/internal/utils"

	corev1 "github.com/flexicky/protos/core.core.v1"
	"google.golang.org/grpc"
)

type server struct {
	corev1.UnimplementedAuthServer

	authAction     authAction.Action
	registerAction registerAction.Action
	meAction       meAction.Action
	validator      *validator.Validator
}

// Register binds the RPC handlers to the gRPC server.
func Register(
	grpcServer *grpc.Server,
	auth authAction.Action,
	register registerAction.Action,
	me meAction.Action,
	validate *validator.Validator,
) {
	corev1.RegisterAuthServer(grpcServer, &server{
		authAction:     auth,
		registerAction: register,
		meAction:       me,
		validator:      validate,
	})
}

func (s *server) Login(ctx context.Context, req *corev1.LoginRequest) (*corev1.LoginResponse, error) {
	params := authDto.Login{Email: req.GetEmail(), Password: req.GetPassword()}

	ctx = withClientInfo(ctx)

	if err := s.validator.Validate(params); err != nil {
		return &corev1.LoginResponse{Ok: false, Message: utils.Ptr(apperr.PublicMessage(apperr.NewValidationError(err.Error())))}, nil
	}

	return s.authAction.Run(ctx, params), nil
}

func (s *server) Register(
	ctx context.Context,
	req *corev1.RegisterRequest,
) (*corev1.RegisterResponse, error) {
	params := userDto.NewUser{Email: req.GetEmail(), Password: req.GetPassword()}

	ctx = withClientInfo(ctx)

	if err := s.validator.Validate(params); err != nil {
		return &corev1.RegisterResponse{Ok: false, Message: utils.Ptr(apperr.PublicMessage(apperr.NewValidationError(err.Error())))}, nil
	}

	return s.registerAction.Run(ctx, params), nil
}

func (s *server) Me(ctx context.Context, _ *corev1.MeRequest) (*corev1.MeResponse, error) {
	ctx = withClientInfo(ctx)

	return s.meAction.Run(ctx), nil
}

// withClientInfo puts the caller's user agent and IP address into the context.
func withClientInfo(ctx context.Context) context.Context {
	return ctxkeys.WithClientInfo(ctx, utils.ClientInfoFromMetadata(ctx))
}
