package depositer

import (
	"context"
	"github.com/tokend/stellar-deposit-svc/internal/config"
	"github.com/tokend/stellar-deposit-svc/internal/horizon"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/getters"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/submit"
	"github.com/tokend/stellar-deposit-svc/internal/services/payment"
	"github.com/tokend/stellar-deposit-svc/internal/services/issuer"
	"github.com/tokend/stellar-deposit-svc/internal/services/watchlist"
	"github.com/tokend/stellar-deposit-svc/internal/transaction"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/running"
	"gitlab.com/tokend/go/xdrbuild"
	"sync"
	"time"
)

type Service struct {
	assetWatcher *watchlist.Service
	log          *logan.Entry
	config       config.Config
	spawned      map[string]bool
	assets       <-chan watchlist.Details
	wg           *sync.WaitGroup
	builder      *xdrbuild.Builder
}

func New(cfg config.Config) *Service {
	wg := &sync.WaitGroup{}
	assetWatcher := watchlist.New(watchlist.Opts{
		AssetOwner: cfg.DepositConfig().AssetOwner.Address(),
		Streamer:   getters.NewDefaultAssetHandler(cfg.Horizon()),
		Log:        cfg.Log(),
		Timeout:    cfg.DepositConfig().Delay,
		Wg:         wg,
	})
	builder, err := horizon.NewConnector(cfg.Horizon()).Builder()
	if err != nil {
		cfg.Log().WithError(err).Fatal("failed to make builder")
	}
	return &Service{
		log:          cfg.Log(),
		config:       cfg,
		wg:           wg,
		assetWatcher: assetWatcher,
		assets:       assetWatcher.GetChan(),
		spawned:      make(map[string]bool),
		builder: builder,
	}
}

func (s *Service) Run(ctx context.Context) {
	go s.assetWatcher.Run(ctx)

	running.WithBackOff(ctx, s.log, "gatherer", func(ctx context.Context) error {
		for asset := range s.assets {
			s.spawn(ctx, asset)
		}
		return nil
	}, s.config.DepositConfig().Delay, s.config.DepositConfig().Delay, 5*time.Minute)

	s.wg.Wait()
}

func (s *Service) spawn(ctx context.Context, details watchlist.Details) {
	if s.spawned[details.Asset.ID] {
		return
	}
	s.wg.Add(2)
	paymentStreamer := payment.NewService(payment.Opts{
		Client:       s.config.Stellar(),
		Delay:        s.config.StellarConfig().Delay,
		Log:          s.log,
		WatchAddress: s.config.StellarConfig().TargetAddress,
		AssetDetails: details,
		WG:           s.wg,
	})

	payments := paymentStreamer.GetChan()

	issuer := issuer.New(issuer.Opts{
		AssetDetails: details,
		Log:          s.log,
		Streamer: transaction.NewStreamer(
			getters.NewDefaultTransactionHandler(s.config.Horizon()),
		),
		Builder:     s.builder,
		Signer: s.config.DepositConfig().AssetIssuer,
		TxSubmitter: submit.New(s.config.Horizon()),
		Ch:          payments,
		WG:          s.wg,
	})
	s.spawned[details.Asset.ID] = true

	go issuer.Run(ctx)
	go paymentStreamer.Run(ctx)

	s.log.WithFields(logan.F{
		"asset_code": details.StellarDetails.Code,
		"asset_type": details.StellarDetails.AssetType,
	}).Info("Started listening for deposits")
}
