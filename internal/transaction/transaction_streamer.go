package transaction

import (
	"context"
	"fmt"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/getters"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/page"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/query"
	regources "gitlab.com/tokend/regources/generated"
)

const (
	streamPageLimit = 100
)

type Streamer struct {
	getters.TransactionGetter
}

func NewStreamer(transactionGetter getters.TransactionGetter) *Streamer {
	return &Streamer{TransactionGetter: transactionGetter}
}

func (s *Streamer) StreamTransactions(ctx context.Context, changeTypes, entryTypes []int,
) (<-chan regources.TransactionResponse, <-chan error) {
	txChan := make(chan regources.TransactionResponse)
	errChan := make(chan error)
	defer close(txChan)
	defer close(errChan)
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

	processedOnPage := make(map[string]bool)
	go func() {
		for {
			if err != nil {
				errChan <- err
				return
			}
			tx := regources.TransactionResponse{}
			for _, transaction := range txPage.Data {
				if _, ok := processedOnPage[transaction.ID]; ok {
					continue
				}
				processedOnPage[transaction.ID] = true

				tx.Data = transaction
				tx.Meta = txPage.Meta

				for _, relation := range transaction.Relationships.LedgerEntryChanges.Data {
					tx.Included.Add(txPage.Included.MustLedgerEntryChange(relation))
				}

				txChan <- tx
			}

			if len(txPage.Data) < streamPageLimit {
				txPage, err = s.Self()
			} else {
				txPage, err = s.Next()
				processedOnPage = make(map[string]bool)
			}
		}
	}()

	return txChan, errChan

}
