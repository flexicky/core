package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"core/internal/app"
	"core/internal/config"
	"core/internal/logger"
	redisAdapter "core/internal/redis"
	"core/internal/storage"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", os.Getenv("CONFIG_PATH"), "path to the configuration file")
	flag.Parse()

	cfg := config.MustLoad(*configPath)

	log := logger.New(cfg.Env).With(slog.String("service", "core"))

	log.Info("starting application", slog.String("env", cfg.Env))

	application, err := app.New(log, appConfig(cfg))
	if err != nil {
		return err
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- application.Run()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(stop)

	select {
	case err := <-serverErr:
		if err != nil {
			return err
		}

		return nil
	case sign := <-stop:
		log.Info("shutdown signal received", slog.String("signal", sign.String()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := application.Shutdown(ctx); err != nil {
		return err
	}

	log.Info("application stopped")

	return nil
}

func appConfig(cfg *config.Config) app.Config {
	return app.Config{
		GRPCPort:        cfg.GRPC.Port,
		AccessTokenTTL:  cfg.AccessTokenTTL,
		SessionTTL:      cfg.SessionTTL,
		SessionCacheTTL: cfg.SessionCacheTTL,
		// Clean up expired sessions twice per session cache lifetime.
		SessionCleanupInterval: cfg.SessionCacheTTL / 2,
		JWTSeed:                cfg.JWTSeed,
		DB: storage.DBConfig{
			Host:     cfg.DB.Host,
			Port:     cfg.DB.Port,
			Name:     cfg.DB.Name,
			User:     cfg.DB.User,
			Password: cfg.DB.Password,
			SSLMode:  cfg.DB.SSLMode,
		},
		Redis: redisAdapter.RedisConfig{
			Host:     cfg.Redis.Host,
			Port:     cfg.Redis.Port,
			Password: cfg.Redis.Password,
			Prefix:   cfg.Redis.Prefix,
		},
	}
}
