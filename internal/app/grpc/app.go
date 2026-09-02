package grpcapp

import (
	"context"
	"core/internal/app/validator"
	authgrpc "core/internal/grpc/auth"
	"core/internal/middleware/jwt"
	"core/internal/repository/session"
	"core/internal/repository/user"
	"core/internal/service/auth"
	redisServ "core/internal/service/redis"
	sessionServ "core/internal/service/session"
	"core/internal/service/token"
	userServ "core/internal/service/user"
	"core/internal/storage"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"log"
	"log/slog"
	"net"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
	validator  *validator.Validator
}

func New(
	Log *slog.Logger,
	port int,
	pgStorage storage.Storage,
	adapter redis.Client,
) (*App, error) {
	userRepository := user.NewUserRepo(&pgStorage)
	userService := userServ.NewUserService(userRepository)
	sessionRepository := session.NewSessionRepo(&pgStorage)
	sessionService := sessionServ.NewSessionService(userRepository, sessionRepository)

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatalf("error generate private/public keys: %v", err)
	}

	tokenService := token.NewTokenService(privateKey, publicKey)

	redisService := redisServ.NewRedisService(adapter)

	authService := auth.NewAuthService(Log, userService, tokenService, sessionService, redisService)

	whiteList := []string{
		"/auth.Auth/Login",
		"/auth.Auth/Register",
	}

	jwtMiddleware := jwt.NewJWTMiddleware(tokenService, whiteList...)

	gRPCServer := grpc.NewServer(
		grpc.UnaryInterceptor(jwtMiddleware.JWTInterceptor()),
	)

	authgrpc.RegisterServerAPI(gRPCServer, userService, authService)

	return &App{
		log:        Log,
		gRPCServer: gRPCServer,
		port:       port,
	}, nil
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	log := a.log.With(slog.String("op", op))

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	go func() {
		log.Info("starting gRPC server", slog.String("addres", l.Addr().String()))
	}()

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop(ctx context.Context) error {
	done := make(chan struct{})

	go func() {
		a.gRPCServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		a.log.Info("grpcApp closed successfully")
		return nil
	case <-ctx.Done():
		a.log.Warn("grpcApp close timed out", "error", ctx.Err())
		return fmt.Errorf("shutdown timeout: %w", ctx.Err())
	}
}
