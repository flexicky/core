package grpcapp

import (
	"context"
	authgrpc "core/internal/grpc/auth"
	"core/internal/repository/user"
	"core/internal/service"
	"core/internal/storage"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(
	Log *slog.Logger,
	port int,
	pgStorage storage.Storage,
) (*App, error) {
	gRPCServer := grpc.NewServer()

	userRepository := user.NewUserRepo(&pgStorage)
	userService := service.NewUserService(userRepository)

	authgrpc.RegisterServerAPI(gRPCServer, userService)

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

	log.Info("starting gRPC server", slog.String("addres", l.Addr().String()))

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
