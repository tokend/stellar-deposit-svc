package config

import (
	"github.com/stellar/go/clients/horizonclient"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

type config struct {
	stellarConfig   StellarConfig
	depositConfig   DepositConfig
	stellar         horizonclient.ClientInterface

	getter kv.Getter
	once   comfig.Once
	Horizoner
}

func (c *config) Stellar() horizonclient.ClientInterface {
	return c.stellar
}

type Config interface {
	DepositConfig() DepositConfig
	StellarConfig() StellarConfig
	Stellar() horizonclient.ClientInterface
	Horizoner
}

func NewConfig(getter kv.Getter) Config {
	return &config{
		getter:    getter,
		Horizoner: NewHorizoner(getter),
	}
}
