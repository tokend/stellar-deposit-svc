package config

import (
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"time"
)

type StellarConfig struct {
	TargetAddress string        `fig:"target_address"`
	Delay         time.Duration `fig:"delay"`
	IsTestNet     bool          `fig:"is_testnet"`
}

func (c *config) StellarConfig() StellarConfig {
	c.once.Do(func() interface{} {
		var result StellarConfig

		err := figure.
			Out(&result).
			With(figure.BaseHooks).
			From(kv.MustGetStringMap(c.getter, "stellar")).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out stellar"))
		}

		c.stellarConfig = result
		return nil
	})

	return c.stellarConfig
}
