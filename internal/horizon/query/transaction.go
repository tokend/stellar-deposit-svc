package query

import (
	"fmt"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/pages"
)

func TransactionList() string {
	return "/v3/transactions"
}

func TransactionByID(id string) string {
	return fmt.Sprintf("/v3/transactions/%s", id)
}


type TransactionFilters struct {
	EntryTypes  []int `filter:"ledger_entry_changes.entry_types"`
	ChangeTypes []int `filter:"ledger_entry_changes.change_types"`
}

type TransactionIncludes struct {
	LedgerEntryChanges bool `include:"ledger_entry_changes"`
}

type TransactionParams struct {
	Includes TransactionIncludes
	Filters TransactionFilters
	PageParams pages.Params
}
