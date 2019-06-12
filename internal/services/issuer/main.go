package issuer

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/submit"
	"github.com/tokend/stellar-deposit-svc/internal/services/payment"
	"github.com/tokend/stellar-deposit-svc/internal/services/watchlist"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
	"gitlab.com/tokend/addrstate"
	"gitlab.com/tokend/go/amount"
	"gitlab.com/tokend/go/hash"
	"gitlab.com/tokend/go/xdrbuild"
	"gitlab.com/tokend/keypair"
	regources "gitlab.com/tokend/regources/generated"
	"sync"
	"time"
)

type details struct {
	TxHash    string `json:"tx_hash"`
	TxMemo    string `json:"tx_memo"`
	From      string `json:"from"`
	PaymentID string `json:"payment_id"`
}

type txSubmitter interface {
	Submit(ctx context.Context, envelope string, waitIngest bool) (*regources.TransactionResponse, error)
}

type transactionStreamer interface {
	StreamTransactions(ctx context.Context, changeTypes, entryTypes []int,
	) (<-chan regources.TransactionListResponse, <-chan error)
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
	owner           keypair.Address
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
	Signer       keypair.Full
	Log          *logan.Entry
	WG           *sync.WaitGroup
	Ch           <-chan payment.Details
}

func New(opts Opts) *Service {

	return &Service{
		asset:       opts.AssetDetails,
		issuer:      opts.Signer,
		streamer:    opts.Streamer,
		builder:     opts.Builder,
		txSubmitter: opts.TxSubmitter,
		log: opts.Log.WithFields(logan.F{
			"asset_code": opts.AssetDetails.ID,
		}),
		owner: keypair.MustParseAddress(opts.AssetDetails.Relationships.Owner.Data.ID),
		ch:    opts.Ch,
		wg:    opts.WG,
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
	s.log.WithField("asset", s.asset.ID).Info("Started issuer service")
	running.WithBackOff(ctx, s.log, "tokend-issuer", func(ctx context.Context) error {

		for pmnt := range s.ch {
			err := s.processPayment(ctx, pmnt)
			if err != nil {
				return errors.Wrap(err, "failed to process payment", logan.F{
					"tx_hash":    pmnt.TxHash,
					"tx_memo":    pmnt.TxMemo,
					"payment_id": pmnt.ID,
				})
			}
		}

		return nil
	}, 10*time.Second, 20*time.Second, time.Minute)
}

func (s *Service) processPayment(ctx context.Context, payment payment.Details) error {
	address := s.addressProvider.ExternalAccountAt(ctx, payment.LedgerCloseTime, s.asset.ExternalSystemType, payment.TxMemo)
	if address == nil {
		s.log.WithFields(logan.F{
			"payment_id": payment.ID,
			"tx_hash":    payment.TxHash,
			"tx_memo":    payment.TxMemo,
		}).Warn("Unable to find valid address to issue tokens to")
		return nil
	}
	balance := s.addressProvider.Balance(ctx, *address, s.asset.ID)
	if balance == nil {
		s.log.WithFields(logan.F{
			"payment_id": payment.ID,
			"tx_hash":    payment.TxHash,
			"tx_memo":    payment.TxMemo,
			"address":    address,
		}).Warn("Unable to find valid balance to issue tokens to")
		return nil
	}

	issueDetails := details{
		TxMemo:    payment.TxMemo,
		TxHash:    payment.TxHash,
		From:      payment.From,
		PaymentID: payment.GetID(),
	}
	detailsbb, err := json.Marshal(issueDetails)
	if err != nil {
		return errors.Wrap(err, "failed to marshal payment details")
	}

	refHash := hash.Hash([]byte(payment.ID))

	reference := hex.EncodeToString(refHash[:])

	amountToIssue := amount.MustParseU(payment.Amount)

	envelope, err := s.builder.Transaction(s.owner).Op(xdrbuild.CreateIssuanceRequest{
		Reference: reference,
		Asset:     s.asset.ID,
		Amount:    amountToIssue,
		Receiver:  *balance,
		Details:   json.RawMessage(detailsbb),
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

func submitEnvelope(ctx context.Context, envelope string, submitter txSubmitter) error {
	_, err := submitter.Submit(ctx, envelope, false)
	if submitFailure, ok := err.(submit.TxFailure); ok {
		if len(submitFailure.OperationResultCodes) == 1 &&
			submitFailure.OperationResultCodes[0] == "op_reference_duplication" {
			return nil
		}
	}
	if err != nil {
		return errors.Wrap(err, "Horizon SubmitResult has error")
	}

	return nil
}
