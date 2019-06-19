package payment

import (
	"github.com/stellar/go/protocols/horizon/operations"
)


type Details struct {
	TxMemo string `json:"tx_memo"`
	TxHash string `json:"tx_hash"`
	operations.Payment
}
