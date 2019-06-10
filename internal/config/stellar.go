package config

import (
	"time"
)

type StellarConfig struct {
	TargetAddress string        `fig:"target_address"`
	Delay         time.Duration `fig:"delay"`
	IsTestNet     bool          `fig:"is_testnet"`
}

func (c *config) StellarConfig() StellarConfig {
	return c.stellarConfig
}
