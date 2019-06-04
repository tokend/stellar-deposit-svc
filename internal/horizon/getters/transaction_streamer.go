package getters

import (
	"context"
	"fmt"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/pages"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/resources/transactions"
	regources "gitlab.com/tokend/regources/generated"
)

func (g *TransactionGetter) StreamTransactions(ctx context.Context, changeTypes, entryTypes []int,
) (<-chan regources.TransactionResponse, <-chan error) {
	txChan := make(chan regources.TransactionResponse)
	errChan := make(chan error)
	defer close(txChan)
	defer close(errChan)
	limit := fmt.Sprintf("%d", streamPageLimit)
	params := transactions.Params{
		Includes: transactions.Includes{
			LedgerEntryChanges: true,
		},
		Filters: transactions.Filters{
			ChangeTypes: changeTypes,
			EntryTypes:  entryTypes,
		},
		PageParams: pages.Params{
			Limit: &limit,
		},
	}
	txPage, err := g.TransactionList(params)

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
				err = g.PageFromLink(txPage.Links.Self, txPage)
			} else {
				err = g.PageFromLink(txPage.Links.Next, txPage)
				processedOnPage = make(map[string]bool)
			}
		}
	}()

	return txChan, errChan

}
