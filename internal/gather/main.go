package payment

import (
	"context"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/running"
	"gitlab.com/tokend/go/xdrbuild"
	"gitlab.com/tokend/stellar-deposit-svc/internal/config"
	"gitlab.com/tokend/stellar-deposit-svc/internal/payment"
	"gitlab.com/tokend/stellar-deposit-svc/internal/submitter"
	"gitlab.com/tokend/stellar-deposit-svc/internal/watchlist"
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

type Opts struct {
	Log     *logan.Entry
	Config  config.Config
	builder *xdrbuild.Builder
}

func NewService(opts Opts) *Service {
	wg := &sync.WaitGroup{}
	assetWatcher := watchlist.NewService(watchlist.Opts{
		AssetOwner: opts.Config.WatchlistConfig().AssetOwner.Address(),
		Streamer:   opts.Config.Horizon(),
		Log:        opts.Log,
		Timeout:    opts.Config.WatchlistConfig().Delay,
		Wg:         wg,
	})
	return &Service{
		log:          opts.Log,
		config:       opts.Config,
		wg:           wg,
		assetWatcher: assetWatcher,
		assets:       assetWatcher.GetChan(),
		spawned:      make(map[string]bool),
	}
}

func (s *Service) Run(ctx context.Context) {
	go s.assetWatcher.Run(ctx)

	running.WithBackOff(ctx, s.log, "gatherer", func(ctx context.Context) error {
		for asset := range s.assets {
			s.spawn(ctx, asset)
		}
		return nil
	}, s.config.WatchlistConfig().Delay, s.config.WatchlistConfig().Delay, 5*time.Minute)

	s.wg.Wait()
}

func (s *Service) spawn(ctx context.Context, details watchlist.Details) {
	if s.spawned[details.Asset.ID] {
		return
	}

	s.wg.Add(2)
	paymentStreamer := payment.NewService(payment.Opts{
		Client:       s.config.Stellar(),
		Delay:        s.config.PaymentConfig().Delay,
		Log:          s.log,
		WatchAddress: s.config.PaymentConfig().TargetAddress,
		AssetDetails: details,
		WG:           s.wg,
	})

	payments := paymentStreamer.GetChan()

	depositer := submitter.NewService(submitter.Opts{
		AssetDetails: details,
		Log:          s.log,
		Streamer:     s.config.Horizon(),
		Builder:      s.builder,
		AssetIssuer:  s.config.DepositConfig().AssetIssuer,
		TxSubmitter:  s.config.Horizon(),
		Ch:           payments,
		WG:           s.wg,
	})
	s.spawned[details.Asset.ID] = true

	go depositer.Run(ctx)
	go paymentStreamer.Run(ctx)
}
