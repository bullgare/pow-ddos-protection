package hashcash

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	ucontracts "github.com/bullgare/pow-ddos-protection/internal/usecase/contracts"
)

const SeedRandomLen = 16

func NewSeedGenerator(seedRandomLen int) SeedGenerator {
	return SeedGenerator{rndLen: seedRandomLen}
}

var _ ucontracts.SeedGenerator = SeedGenerator{}

type SeedGenerator struct {
	rndLen int
}

// FIXME add unit-tests.
func (g SeedGenerator) Generate(identity string, requestTime time.Time) (string, error) {
	randomString, err := generateRandomString(g.rndLen)
	if err != nil {
		return "", fmt.Errorf("generating random string: %w", err)
	}

	return generateSeed(identity, requestTime, randomString), nil
}

// FIXME add unit-tests.
func generateRandomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

// FIXME add unit-tests.
func generateSeed(clientRemoteAddress string, requestTime time.Time, randomString string) string {
	return base64.StdEncoding.EncodeToString([]byte(
		fmt.Sprintf("%s-%d-%s", clientRemoteAddress, requestTime.UnixNano(), randomString),
	))
}
