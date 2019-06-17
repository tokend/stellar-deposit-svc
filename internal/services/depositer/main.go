package depositer

import (
	"context"
	"github.com/tokend/stellar-deposit-svc/internal/config"
	"github.com/tokend/stellar-deposit-svc/internal/horizon"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/getters"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/submit"
	"github.com/tokend/stellar-deposit-svc/internal/services/issuer"
	"github.com/tokend/stellar-deposit-svc/internal/services/payment"
	"github.com/tokend/stellar-deposit-svc/internal/services/watchlist"
	"github.com/tokend/stellar-deposit-svc/internal/transaction"
	"gitlab.com/distributed_lab/logan/v3"
)

type Service struct {
	assetWatcher *watchlist.Service
	log          *logan.Entry
	config       config.Config
	spawned      map[string]bool
	assets       <-chan watchlist.Details
}

func New(cfg config.Config) *Service {
	assetWatcher := watchlist.New(watchlist.Opts{
		AssetOwner: cfg.DepositConfig().AssetOwner.Address(),
		Streamer:   getters.NewDefaultAssetHandler(cfg.Horizon()),
		Log:        cfg.Log(),
	})

	return &Service{
		log:          cfg.Log(),
		config:       cfg,
		assetWatcher: assetWatcher,
		assets:       assetWatcher.GetChan(),
		spawned:      make(map[string]bool),
	}
}

func (s *Service) Run(ctx context.Context) {
	go s.assetWatcher.Run(ctx)

	for asset := range s.assets {
		s.spawn(ctx, asset)
	}
}

func (s *Service) spawn(ctx context.Context, details watchlist.Details) {
	if s.spawned[details.Asset.ID] {
		return
	}
	paymentStreamer := payment.NewService(payment.Opts{
		Client:       s.config.Stellar(),
		Log:          s.log,
		WatchAddress: s.config.PaymentConfig().TargetAddress,
		AssetDetails: details,
	})

	payments := paymentStreamer.GetChan()

	builder, err := horizon.NewConnector(s.config.Horizon()).Builder()
	if err != nil {
		s.log.WithError(err).Fatal("failed to make builder")
	}

	issueSubmitter := issuer.New(issuer.Opts{
		AssetDetails: details,
		Log:          s.log,
		Streamer: transaction.NewStreamer(
			getters.NewDefaultTransactionHandler(s.config.Horizon()),
		),
		Builder:     builder,
		Signer:      s.config.DepositConfig().AssetIssuer,
		TxSubmitter: submit.New(s.config.Horizon()),
		Ch:          payments,
	})
	s.spawned[details.Asset.ID] = true

	go issueSubmitter.Run(ctx)
	go paymentStreamer.Run(ctx)

	s.log.WithFields(logan.F{
		"asset_code": details.Stellar.Code,
		"asset_type": details.Stellar.AssetType,
	}).Info("Started listening for deposits")
}
