package config

import (
	"github.com/stellar/go/clients/horizonclient"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3"
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

func (c *config) Log() *logan.Entry {
	return c.log
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
		log: logan.New().Level(logan.DebugLevel),
	}
}


func (c *config) Stellar() horizonclient.ClientInterface{
	c.once.Do(func() interface{} {
		var result horizonclient.ClientInterface

		conf := c.StellarConfig()
		switch conf.IsTestNet {
		case true:
			result = horizonclient.DefaultTestNetClient
		case false:
			result = horizonclient.DefaultPublicNetClient
		}
		c.stellar = result
		return nil
	})

	return c.stellar
}
