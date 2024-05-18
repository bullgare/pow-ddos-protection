package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/bullgare/pow-ddos-protection/internal/infra/auth/hashcash"
	"github.com/bullgare/pow-ddos-protection/internal/infra/clients/wordofwisdom"
	tclient "github.com/bullgare/pow-ddos-protection/internal/infra/transport/client"
	"github.com/bullgare/pow-ddos-protection/internal/usecase/handlers/client"
)

const envNetworkAddress = "NETWORK_ADDRESS"

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
	lgr = lgr.With("type", "client")

	onError := onErrorFunc(lgr)

	defer func() {
		if err != nil {
			onError(err)
		}
	}()

	address, ok := os.LookupEnv(envNetworkAddress)
	if !ok {
		return fmt.Errorf("env variable %q is required", envNetworkAddress)
	}

	lgr.Info(fmt.Sprintf("client will communicate to %s...", address))

	transClient, err := tclient.New(address)
	if err != nil {
		return fmt.Errorf("creating transport client: %w", err)
	}

	wowClient, err := wordofwisdom.New(transClient)
	if err != nil {
		return fmt.Errorf("creating word of wisdom client: %w", err)
	}

	authGenerator := hashcash.NewAuthorizer(hashcash.BitLen, hashcash.SaltLen)

	clientRunner := client.RunWordOfWisdom(authGenerator, wowClient, onError, shareInfoFunc(lgr))

	clientRunner(ctx)

	return nil
}

func onErrorFunc(lgr *slog.Logger) func(err error) {
	return func(err error) {
		if err != nil {
			lgr.Error(err.Error())
		}
	}
}

func shareInfoFunc(lgr *slog.Logger) func(msg string) {
	return func(msg string) {
		lgr.Info(msg)
	}
}
