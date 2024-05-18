package contracts

import (
	"time"
)

type Config struct {
	DifficultyLevelPercent int
}

type Authorizer interface {
	GenerateSeed(string, time.Time) (string, error)
	GenerateToken(string, Config) (string, error)
	Check(string, Config) bool

	MergeWithConfig(string, Config) string
	ParseConfigFrom(string) (string, Config, error)
}
