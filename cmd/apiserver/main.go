package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/Saaghh/hezzl-hr/internal/apiserver"
	"github.com/Saaghh/hezzl-hr/internal/config"
	"github.com/Saaghh/hezzl-hr/internal/logger"
	"github.com/Saaghh/hezzl-hr/internal/model"
	"github.com/Saaghh/hezzl-hr/internal/service"
	"github.com/Saaghh/hezzl-hr/internal/store/nats"
	"github.com/Saaghh/hezzl-hr/internal/store/pg"
	"github.com/Saaghh/hezzl-hr/internal/store/rdb"
	migrate "github.com/rubenv/sql-migrate"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	cfg := config.New()

	logger.InitLogger(logger.Config{Level: cfg.LogLevel})

	// no error handling for now
	// check https://github.com/uber-go/zap/issues/991
	//nolint: errcheck
	defer zap.L().Sync()

	pgStore, err := pg.New(ctx, cfg)
	if err != nil {
		zap.L().With(zap.Error(err)).Panic("main/pgStore.New")
	}

	if err = pgStore.Migrate(migrate.Up); err != nil {
		zap.L().With(zap.Error(err)).Panic("main/pgStore.Migrate")
	}

	zap.L().Info("successful postgres migration")

	redisCash := rdb.New(cfg)

	natsPublisher, err := nats.NewPublisher()
	if err != nil {
		zap.L().With(zap.Error(err)).Panic("main/nats.NewPublisher()")
	}

	serviceLayer := service.New(pgStore, redisCash, natsPublisher)

	_, err = serviceLayer.CreateProject(ctx, model.Project{Name: "Первая запись"})
	if err != nil {
		zap.L().With(zap.Error(err)).Panic("error creating standard project")
	}

	server := apiserver.New(
		apiserver.Config{BindAddress: cfg.BindAddress},
		serviceLayer)

	// create first project in db

	if err = server.Run(ctx); err != nil {
		zap.L().With(zap.Error(err)).Panic("main/server.Run")
	}
}
