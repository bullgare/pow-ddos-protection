package main

import (
	"github.com/kelseyhightower/envconfig"
)

func newConfig() (config, error) {
	cfg := &config{}
	err := envconfig.Process("", cfg)
	return *cfg, err
}

type config struct {
	NetworkAddress  string  `envconfig:"NETWORK_ADDRESS" required:"true"`
	RedisAddress    string  `envconfig:"REDIS_ADDRESS" required:"true"`
	TargetRPS       float64 `envconfig:"TARGET_RPS" required:"true"`
	InfoLogsEnabled bool    `envconfig:"INFO_LOGS_ENABLED" default:"false"`
}
