package payment

import (
	"github.com/stellar/go/clients/horizonclient"
	"github.com/tokend/stellar-deposit-svc/internal/services/watchlist"
	"gitlab.com/distributed_lab/logan/v3"
)

type Service struct {
	assetType    horizonclient.AssetType
	assetCode    string
	log          *logan.Entry
	watchAddress string
	client       horizonclient.ClientInterface
	ch           chan Details
}

type Opts struct {
	AssetDetails watchlist.Details
	Log          *logan.Entry
	WatchAddress string
	Client       horizonclient.ClientInterface
}

func NewService(opts Opts) *Service {
	ch := make(chan Details, 100)
	return &Service{
		log: opts.Log.WithFields(logan.F{
			"account_address": opts.WatchAddress,
			"asset_type":      opts.AssetDetails.Stellar.AssetType,
			"asset_code":      opts.AssetDetails.Stellar.Code,
		}),
		assetType:    horizonclient.AssetType(opts.AssetDetails.Stellar.AssetType),
		assetCode:    opts.AssetDetails.Stellar.Code,
		watchAddress: opts.WatchAddress,
		client:       opts.Client,
		ch:           ch,
	}
}
