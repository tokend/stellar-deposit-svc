package issuer

import (
	"context"
	"encoding/json"
	"github.com/tokend/stellar-deposit-svc/internal/services/payment"
	"github.com/tokend/stellar-deposit-svc/internal/services/watchlist"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
	"gitlab.com/tokend/addrstate"
	"gitlab.com/tokend/go/amount"
	"gitlab.com/tokend/go/xdrbuild"
	"gitlab.com/tokend/keypair"
	regources "gitlab.com/tokend/regources/generated"
	"sync"
	"time"
)

type txSubmitter interface {
	Submit(ctx context.Context, envelope string) (*regources.TransactionResponse, error)
}

type transactionStreamer interface {
	StreamTransactions(ctx context.Context, changeTypes, entryTypes []int,
	) (<-chan regources.TransactionResponse, <-chan error)
}

// addressProvider must be implemented by WatchAddress storage to pass into Service constructor.
type addressProvider interface {
	ExternalAccountAt(ctx context.Context, ts time.Time, externalSystem int32, externalData string) (address *string)
	Balance(ctx context.Context, address string, asset string) (balance *string)
}

type Service struct {
	streamer        transactionStreamer
	txSubmitter     txSubmitter
	builder         *xdrbuild.Builder
	asset           watchlist.Details
	issuer          keypair.Full
	log             *logan.Entry
	addressProvider addressProvider
	ch              <-chan payment.Details
	wg              *sync.WaitGroup
}

type Opts struct {
	Streamer     transactionStreamer
	TxSubmitter  txSubmitter
	Builder      *xdrbuild.Builder
	AssetDetails watchlist.Details
	AssetIssuer  keypair.Full
	Log          *logan.Entry
	WG           *sync.WaitGroup
	Ch           <-chan payment.Details
}

func New(opts Opts) *Service {

	return &Service{
		asset:       opts.AssetDetails,
		issuer:      opts.AssetIssuer,
		streamer:    opts.Streamer,
		builder:     opts.Builder,
		txSubmitter: opts.TxSubmitter,
		log: opts.Log.WithFields(logan.F{
			"asset_code": opts.AssetDetails.ID,
		}),
		ch: opts.Ch,
		wg: opts.WG,
	}
}

func (s *Service) Run(ctx context.Context) {
	defer s.wg.Done()

	s.addressProvider = addrstate.New(
		ctx,
		s.log,
		[]addrstate.StateMutator{
			addrstate.ExternalSystemBindingMutator{SystemType: s.asset.ExternalSystemType},
			addrstate.BalanceMutator{Asset: s.asset.ID},
		},
		s.streamer,
	)
	running.WithBackOff(ctx, s.log, "tokend-issuer", func(ctx context.Context) error {

		for payment := range s.ch {
			err := s.processPayment(ctx, payment)
			if err != nil {
				return errors.Wrap(err, "failed to process payment", logan.F{
					"tx_hash":    payment.TxHash,
					"tx_memo":    payment.TxMemo,
					"payment_id": payment.ID,
				})
			}
		}

		return nil
	}, 10*time.Second, 20*time.Second, time.Minute)
}

func (s *Service) processPayment(ctx context.Context, payment payment.Details) error {
	address := s.addressProvider.ExternalAccountAt(ctx, payment.LedgerCloseTime, s.asset.ExternalSystemType, payment.TxMemo)
	if address == nil {
		//todo
		return nil
	}
	balance := s.addressProvider.Balance(ctx, *address, s.asset.ID)
	if balance == nil {
		//todo
		return nil
	}
	detailsbb, err := json.Marshal(payment)
	if err != nil {
		return errors.Wrap(err, "failed to marshal payment details")
	}
	tasks := uint32(0)
	envelope, err := s.builder.Transaction(s.issuer).Op(xdrbuild.CreateIssuanceRequest{
		Reference: payment.ID,
		Asset:     s.asset.ID,
		Amount:    amount.MustParseU(payment.Amount),
		Receiver:  *balance,
		Details:   json.RawMessage(detailsbb),
		AllTasks:  &tasks,
	}).Sign(s.issuer).Marshal()
	if err != nil {
		return errors.Wrap(err, "failed to craft transaction")
	}

	err = submitEnvelope(ctx, envelope, s.txSubmitter)
	if err != nil {
		return errors.Wrap(err, "failed to submit issuance tx")
	}

	return nil
}

func submitEnvelope(ctx context.Context, envelope string, submitter txSubmitter) (error) {
	result, err := submitter.Submit(ctx, envelope)
	if err != nil {
		return errors.Wrap(err, "Horizon SubmitResult has error", logan.F{
			"submit_result": result,
		})
	}

	return nil
}
