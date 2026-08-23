package app

import (
	grpcapp "core/internal/app/grpc"
	postgresapp "core/internal/app/postgres"
	postgresStorage "core/internal/storage"
	"log/slog"
	"time"
)

type App struct {
	GRPCServer *grpcapp.App
	PGStorage  *postgresapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	storagePath string,
	tokenTTL time.Duration,
	dbCfg postgresStorage.DBConfig,
) *App {
	grpcApp, err := grpcapp.New(log, grpcPort)
	if err != nil {
		log.Info("error init grpc server %w", err)
	}

	pgApp, err := postgresapp.New(log, dbCfg)
	if err != nil {
		log.Info("error init postgres db %w", err)
	}

	return &App{
		GRPCServer: grpcApp,
		PGStorage:  pgApp,
	}
}
