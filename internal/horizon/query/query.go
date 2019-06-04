package query

import (
	"fmt"
	"net/url"
	"reflect"
)

type Params interface{}

func handleFilters(values *url.Values, t reflect.Type, v reflect.Value) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if name, ok := field.Tag.Lookup("filter"); ok {
			if value := reflect.Indirect(v).Field(i); value.Kind() != reflect.Ptr || !value.IsNil() {
				values.Add(fmt.Sprintf("filter[%s]", name), fmt.Sprintf("%v",reflect.Indirect(value)))
			}
		}
	}
}

func handlePageParams(values *url.Values, t reflect.Type, v reflect.Value) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if name, ok := field.Tag.Lookup("page"); ok {
			if value := reflect.Indirect(v).Field(i); value.Kind() != reflect.Ptr || !value.IsNil() {
				values.Add(fmt.Sprintf("page[%s]", name), fmt.Sprintf("%v",reflect.Indirect(value)))
			}
		}
	}
}

func handleIncludes(values *url.Values, t reflect.Type, v reflect.Value) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if name, ok := field.Tag.Lookup("include"); ok {
			if value := reflect.Indirect(v).Field(i);
				value.Kind() == reflect.Bool && value.Bool() {
				values.Add("include", name)
			}
		}
	}
}

func Prepare(qp Params) (url.Values, error) {
	result := url.Values{}
	if qp == nil {
		return result, nil
	}
	t := reflect.TypeOf(qp)
	v := reflect.ValueOf(qp)
	reflect.New(t)
	if filters, ok := t.FieldByName("Filters"); ok {
		ft := filters.Type
		fv := reflect.Indirect(v).FieldByName("Filters")
		handleFilters(&result, ft, fv)
	}

	if includes, ok := t.FieldByName("Includes"); ok {
		ft := includes.Type
		fv := reflect.Indirect(v).FieldByName("Includes")
		handleIncludes(&result, ft, fv)
	}

	if pageParams, ok := t.FieldByName("PageParams"); ok {
		ft := pageParams.Type
		fv := reflect.Indirect(v).FieldByName("PageParams")
		handlePageParams(&result, ft, fv)

	}

	return result, nil
}
