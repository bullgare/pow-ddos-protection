// Package hashcash is an implementation of an Authorizer contract.
//
// WARNING: The library (github.com/catalinc/hashcash) is not ideal, but it can be easily replaced with a better one later.
package hashcash

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/catalinc/hashcash"

	"github.com/bullgare/pow-ddos-protection/internal/usecase/contracts"
)

const (
	BitsLenMin = 10
	BitsLenMax = 26

	SaltLen = 32
)

const (
	authorizerVersionV1 = "v1"

	separator = ";"
)

func NewAuthorizer(bitsLenMin, bitsLenMax, saltLen uint) Authorizer {
	return Authorizer{
		bitsLenMin: bitsLenMin,
		bitsLenMax: bitsLenMax,
		saltLen:    saltLen,
	}
}

var _ contracts.Authorizer = Authorizer{}

type Authorizer struct {
	bitsLenMin uint
	bitsLenMax uint
	saltLen    uint
}

func (a Authorizer) GenerateToken(uniqueSeed string, cfg contracts.AuthorizerConfig) (string, error) {
	bitsLen := a.calculateBitsLen(cfg)
	token, err := hashcash.New(bitsLen, a.saltLen, authorizerVersionV1).Mint(uniqueSeed)
	if err != nil {
		return "", fmt.Errorf("generating token: %w", err)
	}

	return a.MergeWithConfig(token, cfg), nil
}

func (a Authorizer) Check(token string, cfg contracts.AuthorizerConfig) bool {
	bitsLen := a.calculateBitsLen(cfg)

	return hashcash.New(bitsLen, a.saltLen, authorizerVersionV1).Check(token)
}

func (a Authorizer) MergeWithConfig(data string, cfg contracts.AuthorizerConfig) string {
	return authorizerVersionV1 + separator + strconv.Itoa(cfg.DifficultyLevelPercent) + separator + data
}

func (a Authorizer) ParseConfigFrom(dataWithConfig string) (string, contracts.AuthorizerConfig, error) {
	chunks := strings.SplitN(dataWithConfig, separator, 3)
	if len(chunks) != 3 {
		return "", contracts.AuthorizerConfig{}, fmt.Errorf("got %d chunks in marshalled data with config instead of 3", len(chunks))
	}
	if chunks[0] != authorizerVersionV1 {
		return "", contracts.AuthorizerConfig{}, fmt.Errorf("auth data version %s is not supported", chunks[0])
	}

	level, err := strconv.Atoi(chunks[1])
	if err != nil {
		return "", contracts.AuthorizerConfig{}, fmt.Errorf("parsing difficulty level percent: %w", err)
	}

	return chunks[2], contracts.AuthorizerConfig{DifficultyLevelPercent: level}, nil
}

func (a Authorizer) calculateBitsLen(cfg contracts.AuthorizerConfig) uint {
	level := cfg.DifficultyLevelPercent
	level = max(level, 0)
	level = min(level, 100)

	bitsLen := a.bitsLenMin + (a.bitsLenMax-a.bitsLenMin)*uint(level)/100

	return bitsLen
}
