package getters

//go:generate genny -in=getter.tpl -out=asset_getter.go gen "Template=Asset"
//go:generate genny -in=getter.tpl -out=transaction_getter.go gen "Template=Transaction"

const (
	streamPageLimit = 100
)