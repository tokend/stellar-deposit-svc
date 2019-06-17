package watchlist

import (
	. "github.com/go-ozzo/ozzo-validation"
	"github.com/stellar/go/clients/horizonclient"
	regources "gitlab.com/tokend/regources/generated"
)

var assetTypes = []string{
	string(horizonclient.AssetTypeNative),
	string(horizonclient.AssetType4),
	string(horizonclient.AssetType12),
}

type AssetDetails struct {
	ExternalSystemType int32 `json:"external_system_type,string"`
	Stellar            struct {
		Deposit   bool   `json:"deposit"`
		AssetType string `json:"asset_type"`
		Code      string `json:"asset_code"`
	} `json:"stellar"`
}

func (s AssetDetails) Validate() error {
	errs := Errors{
		"ExternalSystemType": Validate(&s.ExternalSystemType, Required, Min(1)),
		"AssetType":          Validate(&s.Stellar.AssetType, Required, In(assetTypes)),
		"Deposit":            Validate(&s.Stellar, Required),
	}

	if s.Stellar.AssetType == string(horizonclient.AssetType4) {
		errs["Code"] = Validate(&s.Stellar.AssetType, Required, Length(1, 4))
	}

	if s.Stellar.AssetType == string(horizonclient.AssetType12) {
		errs["Code"] = Validate(&s.Stellar.AssetType, Required, Length(5, 12))
	}

	return errs.Filter()
}

type Details struct {
	regources.Asset
	AssetDetails
}
