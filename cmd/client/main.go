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

	lgr.Info(fmt.Sprintf("client is ready to communicate with %s...", address))

	transClient, err := tclient.New(address)
	if err != nil {
		return fmt.Errorf("creating transport client: %w", err)
	}

	wowClient, err := wordofwisdom.New(transClient)
	if err != nil {
		return fmt.Errorf("creating word of wisdom client: %w", err)
	}

	difficultyManager := hashcash.NoOpDifficultyManagerForClient{} // TODO Ideally, there should be 2 constructors and 2 interfaces for authorizer instead (server/client)

	authGenerator := hashcash.NewAuthorizer(hashcash.BitsLenMin, hashcash.BitsLenMax, hashcash.SaltLen, difficultyManager)

	clientRunner := client.RunWordOfWisdom(authGenerator, wowClient, onError, shareInfoFunc(lgr))

	clientRunner(ctx)
	shareInfoFunc(lgr)("CLIENT IS EXITING")

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
