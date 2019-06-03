package transactions

import (
	"fmt"
	"gitlab.com/tokend/stellar-deposit-svc/internal/horizon/pages"
)

const (
	endpoint = "/v3/transactions"
)

func List() string {
	return endpoint
}

func ByID(id string) string {
	return fmt.Sprintf("%s/%s", endpoint, id)
}


type Filters struct {
	EntryTypes  []int `filter:"ledger_entry_changes.entry_types"`
	ChangeTypes []int `filter:"ledger_entry_changes.change_types"`
}

type Includes struct {
	LedgerEntryChanges bool `include:"ledger_entry_changes"`
}

type Params struct {
	Includes Includes
	Filters Filters
	PageParams pages.Params
}

