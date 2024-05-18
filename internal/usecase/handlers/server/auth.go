package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	dcontracts "github.com/bullgare/pow-ddos-protection/internal/domain/contracts"
	ucontracts "github.com/bullgare/pow-ddos-protection/internal/usecase/contracts"
	"github.com/bullgare/pow-ddos-protection/internal/usecase/users"
)

const rndSize = 16

// Auth generates a rule for a client to pass the auth.
func Auth(authStorage dcontracts.AuthStorage) HandlerAuth {
	return func(ctx context.Context, req ucontracts.AuthRequest) (ucontracts.AuthResponse, error) {
		user, ok := users.FromContext(ctx)
		if !ok {
			return ucontracts.AuthResponse{}, errors.New("user data is not provided")
		}

		randomString, err := generateRandomString(rndSize)
		if err != nil {
			return ucontracts.AuthResponse{}, fmt.Errorf("generating random seed: %w", err)
		}

		seed := generateSeed(user.RemoteAddress, user.RequestTime, randomString)

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

func generateRandomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

func generateSeed(clientRemoteAddress string, requestTime time.Time, randomString string) string {
	return base64.StdEncoding.EncodeToString([]byte(
		fmt.Sprintf("%s-%d-%s", clientRemoteAddress, requestTime.UnixNano(), randomString),
	))
}
