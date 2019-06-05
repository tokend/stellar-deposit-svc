package query

import (
	"fmt"
	"net/url"
	"reflect"
)

type Params interface {
	Filter() interface{}
	Include() interface{}
	Page() interface{}
}

func handleFilters(values *url.Values, t reflect.Type, v reflect.Value) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if name, ok := field.Tag.Lookup("filter"); ok {
			if value := reflect.Indirect(v).Field(i); value.Kind() != reflect.Ptr || !value.IsNil() {
				values.Add(fmt.Sprintf("filter[%s]", name), fmt.Sprintf("%v", reflect.Indirect(value)))
			}
		}
	}
}

func handlePageParams(values *url.Values, t reflect.Type, v reflect.Value) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if name, ok := field.Tag.Lookup("page"); ok {
			if value := reflect.Indirect(v).Field(i); value.Kind() != reflect.Ptr || !value.IsNil() {
				values.Add(fmt.Sprintf("page[%s]", name), fmt.Sprintf("%v", reflect.Indirect(value)))
			}
		}
	}
}

func handleIncludes(values *url.Values, t reflect.Type, v reflect.Value) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if name, ok := field.Tag.Lookup("include"); ok {
			if value := reflect.Indirect(v).Field(i); value.Kind() == reflect.Bool && value.Bool() {
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

	if filters := qp.Filter(); filters != nil {
		ft := reflect.TypeOf(filters)
		fv := reflect.ValueOf(filters)
		handleFilters(&result, ft, fv)
	}

	if includes := qp.Include(); includes != nil {
		ft := reflect.TypeOf(includes)
		fv := reflect.ValueOf(includes)
		handleIncludes(&result, ft, fv)
	}

	if pageParams := qp.Page(); pageParams != nil {
		ft := reflect.TypeOf(pageParams)
		fv := reflect.ValueOf(pageParams)
		handlePageParams(&result, ft, fv)

	}

	return result, nil
}
