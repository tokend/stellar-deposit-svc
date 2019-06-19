package transaction

import (
	"context"
	"fmt"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/getters"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/page"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/query"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
	regources "gitlab.com/tokend/regources/generated"
	"time"
)

const (
	streamPageLimit = 100
)

type Streamer struct {
	getters.TransactionHandler
}

func NewStreamer(handler getters.TransactionHandler) *Streamer {
	return &Streamer{handler}
}
func (s *Streamer) StreamTransactions(ctx context.Context, changeTypes, entryTypes []int,
) (<-chan regources.TransactionListResponse, <-chan error) {
	txChan := make(chan regources.TransactionListResponse)
	errChan := make(chan error)
	limit := fmt.Sprintf("%d", streamPageLimit)
	s.SetFilters(query.TransactionFilters{
		ChangeTypes: changeTypes,
		EntryTypes:  entryTypes,
	})
	s.SetPageParams(page.Params{
		Limit: &limit,
	})
	s.SetIncludes(query.TransactionIncludes{
		LedgerEntryChanges: true,
	})

	txPage, err := s.List()

	go func() {
		defer close(txChan)
		defer close(errChan)
		running.WithBackOff(ctx, logan.New(), "tx-streamer", func(ctx context.Context) error {
			if err != nil {
				errChan <- err
				return errors.Wrap(err, "error occurred while streaming transactions")
			}
			if txPage != nil && len(txPage.Data) != 0 {
				txChan <- *txPage
			}

			txPage, err = s.Next()

			return nil
		}, 10*time.Second, 15*time.Second, 5*time.Minute)
	}()

	return txChan, errChan

}
