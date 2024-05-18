package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/bullgare/pow-ddos-protection/internal/app/server"
	"github.com/bullgare/pow-ddos-protection/internal/infra/auth/hashcash"
	"github.com/bullgare/pow-ddos-protection/internal/infra/repositories"
	"github.com/bullgare/pow-ddos-protection/internal/infra/transport/listener"
	handlers "github.com/bullgare/pow-ddos-protection/internal/usecase/handlers/server"
)

func main() {
	// TODO add graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := run(ctx)
	if err != nil {
		os.Exit(1)
	}
}

func run(ctx context.Context) (err error) {
	lgr := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	lgr = lgr.With("type", "server")

	onError := onErrorFunc(lgr)

	defer func() {
		if err != nil {
			onError(err)
		}
	}()

	cfg, err := newConfig()
	if err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}

	lgr.Info(fmt.Sprintf("starting the server on %s...", cfg.NetworkAddress))

	lsn, err := listener.New(cfg.NetworkAddress, onError, shareInfoFunc(cfg.InfoLogsEnabled, lgr))
	if err != nil {
		return fmt.Errorf("creating tcp listener: %w", err)
	}

	authStorage, err := repositories.NewAuthStorage(cfg.RedisAddress)
	if err != nil {
		return fmt.Errorf("creating auth storage: %w", err)
	}

	wowQuotes := repositories.NewWOW()

	seedGenerator := hashcash.NewSeedGenerator(hashcash.SeedRandomLen)

	difficultyManager, stop := hashcash.NewDifficultyManager(cfg.TargetRPS, hashcash.DifficultyChangeStep)
	defer stop()

	authChecker := hashcash.NewAuthorizer(hashcash.BitsLenMin, hashcash.BitsLenMax, hashcash.SaltLen, difficultyManager)

	handlerAuth := handlers.Auth(seedGenerator, authChecker, authStorage, shareInfoFunc(cfg.InfoLogsEnabled, lgr))
	handlerData := handlers.Data(authChecker, authStorage, wowQuotes)

	srv, err := server.New(lsn, handlerAuth, handlerData, onError)
	if err != nil {
		return fmt.Errorf("creating app server: %w", err)
	}
	defer srv.Stop()

	err = srv.Start(ctx)
	if err != nil {
		return fmt.Errorf("starting app server: %w", err)
	}

	return nil
}

func onErrorFunc(lgr *slog.Logger) func(err error) {
	return func(err error) {
		if err != nil {
			lgr.Error(err.Error())
		}
	}
}

func shareInfoFunc(enabled bool, lgr *slog.Logger) func(msg string) {
	return func(msg string) {
		if enabled {
			lgr.Info(msg)
		}
	}
}
