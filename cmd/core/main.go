package main

import (
	"context"
	"core/internal/app"
	"core/internal/config"
	"core/internal/storage"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// TODO crate database core
// TODO gracefullShotdown make
// TODO create migrations

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting application")

	if cfg.Env == envLocal {
		log.Info("cfg data - ", cfg)
	}

	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL, storage.DBConfig{
		Host:     cfg.DB.Host,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		Port:     cfg.DB.Port,
		SSLMode:  cfg.DB.SslMode,
		Name:     cfg.DB.Name,
	})

	go application.GRPCServer.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	log.Info("sttoping application", slog.String("signal", sign.String()))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := application.Shotdown(ctx); err != nil {
		log.Error("graceful shutdown error", "error", err)
	}

	log.Info("Application stopped")
}

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
