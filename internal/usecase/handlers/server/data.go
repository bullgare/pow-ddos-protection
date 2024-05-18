package server

import (
	"context"
	"errors"
	"fmt"

	dcontracts "github.com/bullgare/pow-ddos-protection/internal/domain/contracts"
	ucontracts "github.com/bullgare/pow-ddos-protection/internal/usecase/contracts"
	"github.com/bullgare/pow-ddos-protection/internal/usecase/users"
)

func Data(
	authChecker ucontracts.Authorizer,
	authStorage dcontracts.AuthStorage,
	wowQuotes dcontracts.WOWQuotes,
) HandlerData {
	return func(ctx context.Context, req ucontracts.DataRequest) (ucontracts.DataResponse, error) {
		user, ok := users.FromContext(ctx)
		if !ok {
			return ucontracts.DataResponse{}, errors.New("user not found")
		}

		cacheReq := dcontracts.AuthData{
			Seed:   req.OriginalSeed,
			UserID: user.RemoteAddress,
		}
		exists, err := authStorage.CheckExists(ctx, cacheReq)
		if err != nil {
			return ucontracts.DataResponse{}, fmt.Errorf("authStorage.CheckExists: %w", err)
		}
		if !exists {
			return ucontracts.DataResponse{}, errors.New("user did not request an auth")
		}
		_ = authStorage.Delete(ctx, cacheReq)

		token, cfg, err := authChecker.ParseConfigFrom(req.Token)
		if err != nil {
			return ucontracts.DataResponse{}, fmt.Errorf("parsing config from raw token: %w", err)
		}

		if !authChecker.Check(token, cfg) {
			return ucontracts.DataResponse{}, errors.New("user provided invalid auth token")
		}

		return ucontracts.DataResponse{
			MyPrecious: wowQuotes.GetRandomQuote(),
		}, nil
	}
}
