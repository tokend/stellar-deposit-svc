package config

import (
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (c *config) Log() *logan.Entry {
	c.once.Do(func() interface{} {
		var config struct {
			Level logan.Level `fig:"level,required"`
		}

		err := figure.
			Out(&config).
			With(figure.BaseHooks).
			From(kv.MustGetStringMap(c.getter, "log")).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out horizon"))
		}

		c.log = logan.New().Level(config.Level)
		return nil
	})

	return c.log
}
