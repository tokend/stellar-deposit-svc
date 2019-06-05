package query

import (
	"fmt"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/page"
)

type AssetFilters struct {
	Owner  *string `filter:"owner"`
	Policy *uint32 `filter:"policy"`
}

type AssetIncludes struct {
	Owner bool `include:"owner"`
}

type AssetParams struct {
	Includes   AssetIncludes
	Filters    AssetFilters
	PageParams page.Params
}

func (p AssetParams) Filter() interface{} {
	return p.Filters
}

func (p AssetParams) Include() interface{} {
	return p.Includes
}

func (p AssetParams) Page() interface{} {
	return p.PageParams
}

func AssetByID(code string) string {
	return fmt.Sprintf("/v3/assets/%s", code)
}

func AssetList() string {
	return "/v3/assets"
}
