package config

import (
	"github.com/stellar/go/clients/horizonclient"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

type config struct {
	depositConfig DepositConfig
	paymentConfig PaymentConfig
	stellar       horizonclient.ClientInterface

	comfig.Logger
	getter kv.Getter
	once   comfig.Once
	Horizoner
}

type Config interface {
	DepositConfig() DepositConfig
	PaymentConfig() PaymentConfig
	Stellar() horizonclient.ClientInterface
	comfig.Logger
	Horizoner
}

func NewConfig(getter kv.Getter) Config {
	return &config{
		getter:    getter,
		Horizoner: NewHorizoner(getter),
		Logger:    comfig.NewLogger(getter, comfig.LoggerOpts{}),
	}
}
