package config

import (
	"github.com/stellar/go/clients/horizonclient"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type config struct {
	stellarConfig   StellarConfig
	depositConfig   DepositConfig
	stellar         horizonclient.ClientInterface

	log *logan.Entry
	getter kv.Getter
	once   comfig.Once
	Horizoner
}


type Config interface {
	DepositConfig() DepositConfig
	StellarConfig() StellarConfig
	Stellar() horizonclient.ClientInterface
	Log() *logan.Entry
	Horizoner
}

func NewConfig(getter kv.Getter) Config {
	return &config{
		getter:    getter,
		Horizoner: NewHorizoner(getter),
	}
}


func (c *config) Stellar() horizonclient.ClientInterface{
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
		var cl horizonclient.ClientInterface
		switch result.IsTestNet {
		case true:
			cl = horizonclient.DefaultTestNetClient
		case false:
			cl = horizonclient.DefaultPublicNetClient
		}
		c.stellar = cl

		c.stellarConfig = result
		return nil
	})

	return c.stellar
}
