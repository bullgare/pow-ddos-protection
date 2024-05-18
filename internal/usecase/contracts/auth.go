package contracts

import (
	"time"
)

type AuthorizerConfig struct {
	DifficultyLevelPercent int
}

type Authorizer interface {
	GenerateToken(string, AuthorizerConfig) (string, error)
	Check(string, AuthorizerConfig) bool

	MergeWithConfig(string, AuthorizerConfig) string
	ParseConfigFrom(string) (string, AuthorizerConfig, error)
}

type SeedGenerator interface {
	Generate(string, time.Time) (string, error)
}
