package grpcapp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"time"

	authAction "core/internal/action/auth"
	"core/internal/action/me"
	"core/internal/action/register"
	"core/internal/app/validator"
	authgrpc "core/internal/grpc/auth"
	"core/internal/middleware/jwt"
	"core/internal/repository"
	"core/internal/repository/session"
	"core/internal/repository/user"
	"core/internal/service/auth"
	redisServ "core/internal/service/redis"
	sessionServ "core/internal/service/session"
	"core/internal/service/token"
	userServ "core/internal/service/user"
	"core/internal/worker/session_cleaner"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// publicMethods are the RPCs available without an access token.
var publicMethods = []string{
	"/auth.Auth/Login",
	"/auth.Auth/Register",
}

// Config holds the knobs of the gRPC application.
type Config struct {
	Port                   int
	AccessTokenTTL         time.Duration
	SessionTTL             time.Duration
	SessionCacheTTL        time.Duration
	SessionCleanupInterval time.Duration
	RedisPrefix            string
	// JWTSeed is an optional base64 encoded ed25519 seed. When empty, an
	// ephemeral key pair is generated, so tokens do not survive a restart.
	JWTSeed string
}

type App struct {
	log    *slog.Logger
	server *grpc.Server
	port   int
	// stopWorker cancels the background session cleaner.
	stopWorker context.CancelFunc
}

// New wires the whole object graph of the service.
func New(
	log *slog.Logger,
	cfg Config,
	db repository.Querier,
	redisClient *redis.Client,
) (*App, error) {
	keys, err := signingKeys(cfg.JWTSeed, log)
	if err != nil {
		return nil, err
	}

	users := userServ.New(user.New(db))
	sessions := sessionServ.New(session.New(db))
	tokens := token.New(keys.private, keys.public, cfg.AccessTokenTTL)

	cache := auth.NewSessionCache(
		log,
		redisServ.New(redisClient),
		cfg.RedisPrefix,
		cfg.SessionCacheTTL,
	)

	authService := auth.New(log, users, tokens, sessions, cache, cfg.SessionTTL)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(jwt.New(tokens, publicMethods...).UnaryServerInterceptor()),
	)

	authgrpc.Register(
		server,
		authAction.New(log, authService),
		register.New(log, users, authService),
		me.New(log, authService),
		validator.New(),
	)

	reflection.Register(server)

	workerCtx, stopWorker := context.WithCancel(context.Background())
	go session_cleaner.NewProcess(log, sessions, cfg.SessionCleanupInterval).Run(workerCtx)

	return &App{
		log:        log,
		server:     server,
		port:       cfg.Port,
		stopWorker: stopWorker,
	}, nil
}

// Run serves gRPC until the server is stopped.
func (a *App) Run() error {
	const op = "grpcapp.Run"

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(a.port))
	if err != nil {
		return fmt.Errorf("%s: listen: %w", op, err)
	}

	a.log.Info("gRPC server started", slog.String("address", listener.Addr().String()))

	if err := a.server.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return fmt.Errorf("%s: serve: %w", op, err)
	}

	return nil
}

// Stop gracefully stops the server within the deadline of ctx.
func (a *App) Stop(ctx context.Context) error {
	a.stopWorker()

	done := make(chan struct{})

	go func() {
		a.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		a.log.Info("gRPC server stopped")

		return nil
	case <-ctx.Done():
		a.server.Stop()

		return fmt.Errorf("grpc server shutdown: %w", ctx.Err())
	}
}
