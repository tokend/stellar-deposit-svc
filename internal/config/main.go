package config

import (
	"github.com/stellar/go/clients/horizonclient"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

type config struct {
	paymentConfig   PaymentConfig
	depositConfig   DepositConfig
	watchlistConfig WatchlistConfig
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
	PaymentConfig() PaymentConfig
	WatchlistConfig() WatchlistConfig
	Stellar() horizonclient.ClientInterface
	Horizoner
}

func NewConfig(getter kv.Getter) Config {
	return &config{
		getter:    getter,
		Horizoner: NewHorizoner(getter),
		stellar:   horizonclient.DefaultTestNetClient,
	}
}
