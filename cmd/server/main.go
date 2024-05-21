package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"syscall"
	"time"

	"github.com/skovtunenko/graterm"

	"github.com/bullgare/pow-ddos-protection/internal/app/server"
	"github.com/bullgare/pow-ddos-protection/internal/infra/auth/hashcash"
	"github.com/bullgare/pow-ddos-protection/internal/infra/repositories"
	"github.com/bullgare/pow-ddos-protection/internal/infra/transport/listener"
	handlers "github.com/bullgare/pow-ddos-protection/internal/usecase/handlers/server"
)

func main() {
	terminator, ctx := graterm.NewWithSignals(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	err := run(ctx, terminator)
	if err != nil {
		os.Exit(1)
	}

	if err = terminator.Wait(ctx, 10*time.Second); err != nil {
		log.Printf("graceful termination period was timed out: %v", err)
	}
}

func run(ctx context.Context, terminator *graterm.Terminator) (err error) {
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

	difficultyManager, stopManager := hashcash.NewDifficultyManager(cfg.TargetRPS, hashcash.DifficultyChangeStep)
	terminator.WithOrder(1).Register(5*time.Second, func(ctx context.Context) {
		stopManager()
	})

	authChecker := hashcash.NewAuthorizer(hashcash.BitsLenMin, hashcash.BitsLenMax, hashcash.SaltLen, difficultyManager)

	handlerAuth := handlers.Auth(seedGenerator, authChecker, authStorage, shareInfoFunc(cfg.InfoLogsEnabled, lgr))
	handlerData := handlers.Data(authChecker, authStorage, wowQuotes)

	srv, err := server.New(lsn, handlerAuth, handlerData, onError)
	if err != nil {
		return fmt.Errorf("creating app server: %w", err)
	}

	terminator.WithOrder(2).Register(5*time.Second, func(ctx context.Context) {
		srv.Stop()
	})

	go func() {
		err = srv.Start(ctx)
		if err != nil {
			onError(fmt.Errorf("starting app server: %w", err))
		}
	}()

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
