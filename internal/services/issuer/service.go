package issuer

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/submit"
	"github.com/tokend/stellar-deposit-svc/internal/services/payment"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
	"gitlab.com/tokend/addrstate"
	"gitlab.com/tokend/go/amount"
	"gitlab.com/tokend/go/hash"
	"gitlab.com/tokend/go/xdrbuild"
	"time"
)

func (s *Service) Run(ctx context.Context) {
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
	running.WithBackOff(ctx, s.log, "issuer", func(ctx context.Context) error {

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
		}).Debug("Unable to find valid address to issue tokens to")
		return nil
	}
	balance := s.addressProvider.Balance(ctx, *address, s.asset.ID)
	if balance == nil {
		s.log.WithFields(logan.F{
			"payment_id": payment.ID,
			"tx_hash":    payment.TxHash,
			"tx_memo":    payment.TxMemo,
			"address":    address,
		}).Debug("Unable to find valid balance to issue tokens to")
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
	err = s.submitEnvelope(ctx, envelope, payment.GetID())
	if err != nil {
		return errors.Wrap(err, "failed to submit issuance tx")
	}
	return nil
}

func (s *Service) submitEnvelope(ctx context.Context, envelope string, paymentID string) error {
	_, err := s.txSubmitter.Submit(ctx, envelope, false)
	if submitFailure, ok := err.(submit.TxFailure); ok {
		if len(submitFailure.OperationResultCodes) == 1 &&
			submitFailure.OperationResultCodes[0] == "op_reference_duplication" {
			return nil
		}
	}
	if err != nil {
		return errors.Wrap(err, "Horizon SubmitResult has error")
	}
	s.log.WithField("payment_id", paymentID).Info("Successfully processed deposit")
	return nil
}
