package horizon

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/tokend/stellar-deposit-svc/internal/horizon/assets"
	"net/url"
	"reflect"
	"testing"
)

var (
	ownerAddress = "GBA4EX43M25UPV4WIE6RRMQOFTWXZZRIPFAI5VPY6Z2ZVVXVWZ6NEOOB"
	policies = uint32(42)
	params = assets.Params{
		Includes: assets.Includes{
			Owner: true,
		},
		Filters: assets.Filters{
			Owner: &ownerAddress,
			Policy: &policies,
		},
	}

	result = map[string]string{}
	values = url.Values{}
)

func BenchmarkPrepare(b *testing.B) {
	val := url.Values{}
	for n := 0; n < b.N; n++ {
		val, _  = prepareQuery(params)
	}
	values = val
}


func TestPrepareQuery(t *testing.T) {
	t.Run("all params preset", func(t *testing.T) {
		vals, err := prepareQuery(params)
		ty := reflect.TypeOf(params)
		fmt.Println(ty)
		fmt.Println(ty.NumField())
		expectedValue := url.Values{}
		expectedValue.Add("filter[owner]", ownerAddress)
		expectedValue.Add("filter[policy]", fmt.Sprintf("%d", policies))
		expectedValue.Add("include", "owner")
		assert.NoError(t, err)
		assert.Equal(t, expectedValue, vals)
	})
}