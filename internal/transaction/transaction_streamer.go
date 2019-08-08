package transaction

import (
	"context"
	"fmt"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/getters"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/page"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/query"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/tokend/regources/generated"
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
	if err != nil {
		errChan <- err
		return txChan, errChan
	}
	if txPage == nil {
		errChan <- errors.New("got nil page")
		return txChan, errChan
	}
	go func() {
		defer close(txChan)
		defer close(errChan)
		txChan <- *txPage
		ticker := time.NewTicker(5 * time.Second)
		for {
			if len(txPage.Data) == 0 {
				// TODO: Find better way
				<-ticker.C
				txPage, err = s.Self()
			} else {
				txPage, err = s.Next()
			}
			if err != nil {
				errChan <- err
				continue
			}
			if txPage != nil {
				txChan <- *txPage
			}
		}
	}()

	return txChan, errChan

}
