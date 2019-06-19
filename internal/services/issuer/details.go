package issuer


type details struct {
	TxHash    string `json:"tx_hash"`
	TxMemo    string `json:"tx_memo"`
	From      string `json:"from"`
	PaymentID string `json:"payment_id"`
}

