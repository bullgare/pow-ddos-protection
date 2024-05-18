// Package hashcash is an implementation of an Authorizer contract.
//
// WARNING: The library (github.com/catalinc/hashcash) is not ideal, but it can be easily replaced with a better one later.
package hashcash

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/catalinc/hashcash"

	"github.com/bullgare/pow-ddos-protection/internal/usecase/contracts"
)

const (
	SeedRandomLen = 16

	BitsLenMin = 10
	BitsLenMax = 26

	SaltLen = 32
)

const (
	authorizerVersionV1 = "v1"

	separator = ";"
)

func NewAuthorizer(seedRandomLen int, bitsLenMin, bitsLenMax, saltLen uint) Authorizer {
	return Authorizer{
		rndLen:     seedRandomLen,
		bitsLenMin: bitsLenMin,
		bitsLenMax: bitsLenMax,
		saltLen:    saltLen,
	}
}

var _ contracts.Authorizer = Authorizer{}

type Authorizer struct {
	rndLen     int
	bitsLenMin uint
	bitsLenMax uint
	saltLen    uint
}

func (a Authorizer) GenerateSeed(identity string, requestTime time.Time) (string, error) {
	randomString, err := generateRandomString(a.rndLen)
	if err != nil {
		return "", fmt.Errorf("generating random string: %w", err)
	}

	return generateSeed(identity, requestTime, randomString), nil
}

func (a Authorizer) GenerateToken(uniqueSeed string, cfg contracts.Config) (string, error) {
	bitsLen := a.calculateBitsLen(cfg)
	token, err := hashcash.New(bitsLen, a.saltLen, authorizerVersionV1).Mint(uniqueSeed)
	if err != nil {
		return "", fmt.Errorf("generating token: %w", err)
	}

	return a.MergeWithConfig(token, cfg), nil
}

func (a Authorizer) Check(token string, cfg contracts.Config) bool {
	bitsLen := a.calculateBitsLen(cfg)

	return hashcash.New(bitsLen, a.saltLen, authorizerVersionV1).Check(token)
}

func (a Authorizer) MergeWithConfig(data string, cfg contracts.Config) string {
	return authorizerVersionV1 + separator + strconv.Itoa(cfg.DifficultyLevelPercent) + separator + data
}

func (a Authorizer) ParseConfigFrom(dataWithConfig string) (string, contracts.Config, error) {
	chunks := strings.SplitN(dataWithConfig, separator, 3)
	if len(chunks) != 3 {
		return "", contracts.Config{}, fmt.Errorf("got %d chunks in marshalled data with config instead of 3", len(chunks))
	}
	if chunks[0] != authorizerVersionV1 {
		return "", contracts.Config{}, fmt.Errorf("auth data version %s is not supported", chunks[0])
	}

	level, err := strconv.Atoi(chunks[1])
	if err != nil {
		return "", contracts.Config{}, fmt.Errorf("parsing difficulty level percent: %w", err)
	}

	return chunks[2], contracts.Config{DifficultyLevelPercent: level}, nil
}

func (a Authorizer) calculateBitsLen(cfg contracts.Config) uint {
	level := cfg.DifficultyLevelPercent
	level = max(level, 0)
	level = min(level, 100)

	bitsLen := a.bitsLenMin + (a.bitsLenMax-a.bitsLenMin)*uint(level)/100

	return bitsLen
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