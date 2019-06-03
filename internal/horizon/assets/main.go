package assets

import (
	"fmt"
	"gitlab.com/tokend/stellar-deposit-svc/internal/horizon/pages"
)

const (
	endpoint = "/v3/assets"
)

type Filters struct {
	Owner  *string `filter:"owner"`
	Policy *uint32 `filter:"policy"`
}

type Includes struct {
	Owner bool `include:"owner"`
}

type Params struct {
	Includes Includes
	Filters Filters
	PageParams pages.Params
}

func ByID(code string) string {
	return fmt.Sprintf("%s/%s", endpoint, code)
}

func List() string {
	return endpoint
}
