package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/bullgare/pow-ddos-protection/internal/app/server"
	"github.com/bullgare/pow-ddos-protection/internal/infra/transport/connection"
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

	handlerAuth := handlers.Auth()
	handlerData := handlers.Data()

	// FIXME try rearranging dependencies here
	srv, err := server.New(handlerAuth, handlerData, onError)
	if err != nil {
		return fmt.Errorf("creating app server: %w", err)
	}

	connHandler, err := connection.New(srv, onError)
	if err != nil {
		return fmt.Errorf("creating connection handler: %w", err)
	}

	lsn, err := listener.New(address, connHandler, onError)
	if err != nil {
		return fmt.Errorf("creating tcp listener: %w", err)
	}
	defer lsn.Stop()

	err = lsn.Start(ctx)
	if err != nil {
		return fmt.Errorf("starting tcp listener: %w", err)
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
