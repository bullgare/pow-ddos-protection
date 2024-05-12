package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/bullgare/pow-ddos-protection/internal/app/server"
	"github.com/bullgare/pow-ddos-protection/internal/infra/repositories"
	"github.com/bullgare/pow-ddos-protection/internal/infra/transport/listener"
	handlers "github.com/bullgare/pow-ddos-protection/internal/usecase/handlers/server"
)

const envListenerAddress = "LISTENER_ADDRESS"

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

	onError := onErrorFunc(lgr)

	defer func() {
		if err != nil {
			onError(err)
		}
	}()

	address, ok := os.LookupEnv(envListenerAddress)
	if !ok {
		return fmt.Errorf("env variable %q is required", envListenerAddress)
	}

	lgr.Info(fmt.Sprintf("starting the server on %s...", address))

	lsn, err := listener.New(address, onError)
	if err != nil {
		return fmt.Errorf("creating tcp listener: %w", err)
	}

	wowQuotes := repositories.NewWOW()

	handlerAuth := handlers.Auth()
	handlerData := handlers.Data(wowQuotes)

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
