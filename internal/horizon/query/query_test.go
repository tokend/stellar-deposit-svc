package query

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/url"
	"reflect"
	"testing"
)

var (
	ownerAddress = "GBA4EX43M25UPV4WIE6RRMQOFTWXZZRIPFAI5VPY6Z2ZVVXVWZ6NEOOB"
	policies = uint32(42)
	params = AssetParams{
		Includes: AssetIncludes{
			Owner: true,
		},
		Filters: AssetFilters{
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
		val, _  = Prepare(params)
	}
	values = val
}


func TestPrepareQuery(t *testing.T) {
	t.Run("all params preset", func(t *testing.T) {
		vals, err := Prepare(params)
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