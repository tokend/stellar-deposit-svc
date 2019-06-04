package payment

import (
	"context"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/protocols/horizon/operations"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
	"github.com/tokend/stellar-deposit-svc/internal/watchlist"
	"sync"
	"time"
)

type Details struct {
	TxMemo string `json:"tx_memo"`
	TxHash string `json:"tx_hash"`
	operations.Payment
}

type Service struct {
	assetType    horizonclient.AssetType
	assetCode    string
	log          *logan.Entry
	watchAddress string
	client       horizonclient.ClientInterface
	delay        time.Duration
	ch           chan Details
	wg           *sync.WaitGroup
}

type Opts struct {
	AssetDetails watchlist.Details
	Log          *logan.Entry
	WatchAddress string
	Client       horizonclient.ClientInterface
	Delay        time.Duration
	WG           *sync.WaitGroup
}

func NewService(opts Opts) *Service {
	ch := make(chan Details, 100)
	return &Service{
		log: opts.Log.WithFields(logan.F{
			"account_address": opts.WatchAddress,
			"asset_type":      opts.AssetDetails.StellarDetails.AssetType,
			"asset_code":      opts.AssetDetails.StellarDetails.Code,
		}),
		assetType:    horizonclient.AssetType(opts.AssetDetails.StellarDetails.AssetType),
		assetCode:    opts.AssetDetails.StellarDetails.Code,
		watchAddress: opts.WatchAddress,
		delay:        opts.Delay,
		client:       opts.Client,
		ch:           ch,
		wg:           opts.WG,
	}
}

func (s *Service) GetChan() <-chan Details {
	return s.ch
}

func (s *Service) Run(ctx context.Context) {
	defer s.wg.Done()
	request := horizonclient.OperationRequest{
		ForAccount: s.watchAddress,
		Order:      horizonclient.OrderAsc,
	}
	page, err := s.client.Operations(request)
	if err != nil {
		s.log.WithError(err).Error("failed to get payments page")
		return
	}
	running.WithBackOff(ctx, s.log, "stellar-payment-listener", func(ctx context.Context) error {
		for _, record := range page.Embedded.Records {
			payment := record.(operations.Payment)
			tx, err := s.client.TransactionDetail(record.GetTransactionHash())
			if err != nil {
				return errors.Wrap(err, "failed to get parent transaction of payment", logan.F{
					"tx_hash":    record.GetTransactionHash(),
					"payment_id": record.GetID(),
				})
			}
			s.ch <- paymentDetails(payment, tx)
		}

		page, err = s.client.NextOperationsPage(page)
		if err != nil {
			return errors.Wrap(err, "failed to get next page of payments")
		}

		return nil
	}, s.delay, s.delay, 5*time.Minute)
}

func paymentDetails(record operations.Payment, tx horizon.Transaction) Details {
	return Details{
		Payment: record,
		TxHash:  tx.Hash,
		TxMemo:  tx.Memo,
	}
}
