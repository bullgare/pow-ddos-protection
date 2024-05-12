package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/bullgare/pow-ddos-protection/internal/usecase/handlers/contract"
)

const rndSize = 16

// Auth generates a rule for a client to pass the auth.
func Auth() HandlerAuth {
	return func(ctx context.Context, req contract.AuthRequest) (contract.AuthResponse, error) {
		randomString, err := generateRandomString(rndSize)
		if err != nil {
			return contract.AuthResponse{}, fmt.Errorf("generating random seed: %w", err)
		}

		seed := generateSeed(req.ClientRemoteAddress, req.RequestTime, randomString)

		return contract.AuthResponse{
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

	return base64.URLEncoding.EncodeToString(b), nil
}

func generateSeed(clientRemoteAddress string, requestTime time.Time, randomString string) string {
	return base64.URLEncoding.EncodeToString([]byte(
		fmt.Sprintf("%s-%d-%s", clientRemoteAddress, requestTime.UnixNano(), randomString),
	))
}
