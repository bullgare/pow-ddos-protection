// Package hashcash is an implementation of an Authorizer contract.
//
// WARNING: The library (github.com/catalinc/hashcash) is not ideal, but it can be easily replaced with a better one later.
package hashcash

import (
	"fmt"
	"strings"

	"github.com/catalinc/hashcash"

	"github.com/bullgare/pow-ddos-protection/internal/usecase/contracts"
)

const (
	BitLen  = 24
	SaltLen = 32
)

const (
	authorizerVersionV1 = "v1"

	versionSeparator = ":"
)

func NewAuthorizer(bitsLen, saltLen uint) Authorizer {
	return Authorizer{
		bitsLen: bitsLen,
		saltLen: saltLen,
	}
}

var _ contracts.Authorizer = Authorizer{}

type Authorizer struct {
	bitsLen uint
	saltLen uint
}

func (a Authorizer) Generate(uniqueSeed string) (string, error) {
	token, err := hashcash.New(a.bitsLen, a.saltLen, authorizerVersionV1).Mint(uniqueSeed)
	if err != nil {
		return "", fmt.Errorf("generating token: %w", err)
	}

	return authorizerVersionV1 + versionSeparator + token, nil
}

func (a Authorizer) Check(token string) bool {
	chunks := strings.SplitN(token, versionSeparator, 2)
	if len(chunks) != 2 {
		return false
	}
	if chunks[0] != authorizerVersionV1 {
		return false
	}

	return hashcash.New(a.bitsLen, a.saltLen, authorizerVersionV1).Check(chunks[1])
}
