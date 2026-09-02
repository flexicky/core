package app

import (
	"context"
	grpcapp "core/internal/app/grpc"
	postgresapp "core/internal/app/postgres"
	redisApp "core/internal/app/redis"
	"core/internal/redis"
	postgresStorage "core/internal/storage"
	"log/slog"
	"time"
)

type App struct {
	GRPCServer *grpcapp.App
	PGStorage  *postgresapp.App
	redisPul   *redisApp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	storagePath string,
	tokenTTL time.Duration,
	dbCfg postgresStorage.DBConfig,
	redisCfg redis.RedisConfig,
) *App {

	pgApp, err := postgresapp.New(log, dbCfg)
	if err != nil {
		log.Error("error init postgres db %w", err)
	}

	redisApplication := redisApp.New(log, redisCfg)

	grpcApp, err := grpcapp.New(log, grpcPort, pgApp.GetStorage(), *redisApplication.GetClient())
	if err != nil {
		log.Error("error init grpc server %w", err)
	}

	return &App{
		GRPCServer: grpcApp,
		PGStorage:  pgApp,
		redisPul:   redisApplication,
	}
}

func (a *App) Shotdown(ctx context.Context) error {
	if err := a.GRPCServer.Stop(ctx); err != nil {
		return err
	}
	if err := a.PGStorage.Stop(ctx); err != nil {
		return err
	}
	if err := a.redisPul.Stop(ctx); err != nil {
		return err
	}

	return nil
}
