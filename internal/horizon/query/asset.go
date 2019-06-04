package query

import (
	"fmt"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/pages"
)

type AssetFilters struct {
	Owner  *string `filter:"owner"`
	Policy *uint32 `filter:"policy"`
}

type AssetIncludes struct {
	Owner bool `include:"owner"`
}

type AssetParams struct {
	Includes AssetIncludes
	Filters AssetFilters
	PageParams pages.Params
}

func AssetByID(code string) string {
	return fmt.Sprintf("/v3/assets/%s", code)
}

func AssetList() string {
	return "/v3/assets"
}
