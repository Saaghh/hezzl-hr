package main

import (
	"context"
	"fmt"
	"net/url"
	"os/signal"
	"syscall"

	"github.com/Saaghh/hezzl-hr/internal/chlogger/config"
	"github.com/Saaghh/hezzl-hr/internal/chlogger/nats"
	"github.com/Saaghh/hezzl-hr/internal/chlogger/store"
	"github.com/Saaghh/hezzl-hr/internal/logger"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	cfg := config.New()

	logger.InitLogger(logger.Config{Level: cfg.LogLevel})

	// no error handling for now
	// check https://github.com/uber-go/zap/issues/991
	//nolint: errcheck
	defer zap.L().Sync()

	ch, err := store.New(ctx, store.Config{
		BindAddr: cfg.CHBindAddr,
		Database: cfg.CHDatabase,
		Username: cfg.CHUsername,
		Password: cfg.CHPassword,
	})
	if err != nil {
		zap.L().With(zap.Error(err)).Panic("main/store.New(ctx, nil)")
	}

	if err = ch.Migrate(); err != nil {
		zap.L().With(zap.Error(err)).Panic("main/ch.Migrate(migrate.Up)")
	}

	natsBindAddr := url.URL{
		Scheme: "nats",
		Host:   fmt.Sprintf("%s:%s", cfg.NatsHost, cfg.NatsPort),
	}

	zap.L().Debug(natsBindAddr.String())

	sub, err := nats.NewGoodsEventSubscriber(ch, cfg.CHBatchSize, natsBindAddr.String())
	if err != nil {
		zap.L().With(zap.Error(err)).Panic("nats.NewGoodsEventSubscriber(ctx, ch)")
	}

	_, err = sub.SubscribeEventLogger("goods_logs")
	if err != nil {
		zap.L().With(zap.Error(err)).Panic("sub.SubscribeEventLogger(\"goods_logs\")")
	}

	<-ctx.Done()
}
