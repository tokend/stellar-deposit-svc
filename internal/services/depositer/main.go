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
	"gitlab.com/tokend/go/xdrbuild"
	"sync"
)

type Service struct {
	assetWatcher   *watchlist.Service
	log            *logan.Entry
	config         config.Config
	builder        xdrbuild.Builder
	spawned        sync.Map
	assetsToAdd    <-chan watchlist.Details
	assetsToRemove <-chan string
	sync.WaitGroup
}

func New(cfg config.Config) *Service {
	assetWatcher := watchlist.New(watchlist.Opts{
		AssetOwner: cfg.DepositConfig().AssetOwner.Address(),
		Streamer:   getters.NewDefaultAssetHandler(cfg.Horizon()),
		Log:        cfg.Log(),
	})
	builder, err := horizon.NewConnector(cfg.Horizon()).Builder()
	if err != nil {
		cfg.Log().WithError(err).Fatal("failed to make builder")
	}

	return &Service{
		log:     cfg.Log(),
		config:  cfg,
		builder: *builder,

		assetWatcher:   assetWatcher,
		assetsToAdd:    assetWatcher.GetToAdd(),
		assetsToRemove: assetWatcher.GetToRemove(),
		spawned:        sync.Map{},
		WaitGroup:      sync.WaitGroup{},
	}
}

func (s *Service) Run(ctx context.Context) {
	go s.assetWatcher.Run(ctx)

	s.Add(2)
	go s.spawner(ctx)
	go s.cancellor(ctx)
	s.Wait()
}

func (s *Service) spawner(ctx context.Context) {
	defer s.Done()
	for asset := range s.assetsToAdd {
		if _, ok := s.spawned.Load(asset.ID); !ok {
			s.spawn(ctx, asset)
		}
	}
}

func (s *Service) cancellor(ctx context.Context) {
	defer s.Done()
	for asset := range s.assetsToRemove {
		if raw, ok := s.spawned.Load(asset); ok {
			cancelFunc := raw.(context.CancelFunc)
			cancelFunc()
			s.spawned.Delete(asset)
		}
	}
}

func (s *Service) spawn(ctx context.Context, details watchlist.Details) {

	paymentStreamer := payment.NewService(payment.Opts{
		Client:       s.config.Stellar(),
		Log:          s.log.WithField("asset", details.ID),
		WatchAddress: s.config.PaymentConfig().TargetAddress,
		AssetDetails: details,
	})

	payments := paymentStreamer.GetChan()

	issueSubmitter := issuer.New(issuer.Opts{
		AssetDetails: details,
		Log:          s.log.WithField("asset", details.ID),
		Streamer: transaction.NewStreamer(
			getters.NewDefaultTransactionHandler(s.config.Horizon()),
			s.log.WithField("service", "transaction-streamer"),
		),
		Builder:     s.builder,
		Signer:      s.config.DepositConfig().AssetIssuer,
		TxSubmitter: submit.New(s.config.Horizon()),
		Ch:          payments,
	})
	localCtx, cancelFunc := context.WithCancel(ctx)
	s.spawned.Store(details.Asset.ID, cancelFunc)

	go issueSubmitter.Run(localCtx)
	go paymentStreamer.Run(localCtx)

	s.log.WithFields(logan.F{
		"asset_code": details.Stellar.Code,
		"asset_type": details.Stellar.AssetType,
	}).Info("Started listening for deposits")
}
