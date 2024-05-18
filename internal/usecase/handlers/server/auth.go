package server

import (
	"context"
	"errors"
	"fmt"

	dcontracts "github.com/bullgare/pow-ddos-protection/internal/domain/contracts"
	ucontracts "github.com/bullgare/pow-ddos-protection/internal/usecase/contracts"
	"github.com/bullgare/pow-ddos-protection/internal/usecase/users"
)

// Auth generates a rule for a client to pass the auth.
func Auth(
	seedGenerator ucontracts.SeedGenerator,
	authChecker ucontracts.Authorizer,
	authStorage dcontracts.AuthStorage,
) HandlerAuth {
	return func(ctx context.Context, req ucontracts.AuthRequest) (ucontracts.AuthResponse, error) {
		user, ok := users.FromContext(ctx)
		if !ok {
			return ucontracts.AuthResponse{}, errors.New("user data is not provided")
		}

		seed, err := seedGenerator.Generate(user.RemoteAddress, user.RequestTime)
		if err != nil {
			return ucontracts.AuthResponse{}, fmt.Errorf("generating seed: %w", err)
		}

		// FIXME level should come from a rate limiter
		seed = authChecker.MergeWithConfig(seed, ucontracts.AuthorizerConfig{DifficultyLevelPercent: 70})

		cacheReq := dcontracts.AuthData{
			Seed:   seed,
			UserID: user.RemoteAddress,
		}
		err = authStorage.Store(ctx, cacheReq)
		if err != nil {
			return ucontracts.AuthResponse{}, fmt.Errorf("authStorage.Store: %w", err)
		}

		return ucontracts.AuthResponse{
			Seed: seed,
		}, nil
	}
}
