package config

import (
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/tokend/keypair"
	"gitlab.com/tokend/keypair/figurekeypair"
	"time"
)

type WatchlistConfig struct {
	AssetOwner keypair.Address `fig:"asset_owner"`
	Delay      time.Duration   `fig:"delay"`
}

func (c *config) WatchlistConfig() WatchlistConfig {
	c.once.Do(func() interface{} {
		var result WatchlistConfig

		err := figure.
			Out(&result).
			With(figure.BaseHooks, figurekeypair.Hooks).
			From(kv.MustGetStringMap(c.getter, "watchlist")).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out Watchlist"))
		}

		c.watchlistConfig = result
		return nil
	})

	return c.watchlistConfig
}
