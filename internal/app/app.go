package app

import (
	"context"
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

	pgApp, err := postgresapp.New(log, dbCfg)
	if err != nil {
		log.Info("error init postgres db %w", err)
	}

	grpcApp, err := grpcapp.New(log, grpcPort, pgApp.GetStorage())
	if err != nil {
		log.Info("error init grpc server %w", err)
	}

	return &App{
		GRPCServer: grpcApp,
		PGStorage:  pgApp,
	}
}

func (a *App) Shotdown(ctx context.Context) error {
	if err := a.GRPCServer.Stop(ctx); err != nil {
		return err
	}
	if err := a.PGStorage.Stop(ctx); err != nil {
		return err
	}

	return nil
}
