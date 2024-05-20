// Package hashcash is an implementation of an Authorizer contract.
//
// WARNING: The library (github.com/catalinc/hashcash) is not ideal, but it can be easily replaced with a better one later.
package hashcash

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/catalinc/hashcash"

	ucontracts "github.com/bullgare/pow-ddos-protection/internal/usecase/contracts"
)

const (
	BitsLenMin = 10
	BitsLenMax = 30

	SaltLen = 32
)

const (
	authorizerVersionV1 = "v1"

	separator = ";"
)

func NewAuthorizer(bitsLenMin, bitsLenMax, saltLen uint, difficultyManager ucontracts.DifficultyManager) Authorizer {
	return Authorizer{
		difficultyManager: difficultyManager,
		bitsLenMin:        bitsLenMin,
		bitsLenMax:        bitsLenMax,
		saltLen:           saltLen,
	}
}

var _ ucontracts.Authorizer = Authorizer{}

type Authorizer struct {
	difficultyManager ucontracts.DifficultyManager
	bitsLenMin        uint
	bitsLenMax        uint
	saltLen           uint
}

func (a Authorizer) GenerateToken(uniqueSeed string, cfg ucontracts.AuthorizerConfig) (string, error) {
	bitsLen := a.calculateBitsLen(cfg)
	token, err := hashcash.New(bitsLen, a.saltLen, authorizerVersionV1).Mint(uniqueSeed)
	if err != nil {
		return "", fmt.Errorf("generating token: %w", err)
	}

	return a.MergeWithConfig(token, cfg), nil
}

func (a Authorizer) Check(token string, cfg ucontracts.AuthorizerConfig) bool {
	// we only protect functionality guarded by these checks
	a.difficultyManager.IncrRequests()

	bitsLen := a.calculateBitsLen(cfg)

	return hashcash.New(bitsLen, a.saltLen, authorizerVersionV1).Check(token)
}

func (a Authorizer) GenerateConfig() ucontracts.AuthorizerConfig {
	difficulty := a.difficultyManager.GetDifficultyPercent()

	return ucontracts.AuthorizerConfig{DifficultyLevelPercent: difficulty}
}

func (a Authorizer) MergeWithConfig(data string, cfg ucontracts.AuthorizerConfig) string {
	return authorizerVersionV1 + separator + strconv.Itoa(cfg.DifficultyLevelPercent) + separator + data
}

func (a Authorizer) ParseConfigFrom(dataWithConfig string) (string, ucontracts.AuthorizerConfig, error) {
	chunks := strings.SplitN(dataWithConfig, separator, 3)
	if len(chunks) != 3 {
		return "", ucontracts.AuthorizerConfig{}, fmt.Errorf("got %d chunks in marshalled data with config instead of 3", len(chunks))
	}
	if chunks[0] != authorizerVersionV1 {
		return "", ucontracts.AuthorizerConfig{}, fmt.Errorf("auth data version %s is not supported", chunks[0])
	}

	level, err := strconv.Atoi(chunks[1])
	if err != nil {
		return "", ucontracts.AuthorizerConfig{}, fmt.Errorf("parsing difficulty level percent: %w", err)
	}

	return chunks[2], ucontracts.AuthorizerConfig{DifficultyLevelPercent: level}, nil
}

func (a Authorizer) calculateBitsLen(cfg ucontracts.AuthorizerConfig) uint {
	level := cfg.DifficultyLevelPercent
	level = max(level, 0)
	level = min(level, 100)

	bitsLen := a.bitsLenMin + (a.bitsLenMax-a.bitsLenMin)*uint(level)/100

	return bitsLen
}
