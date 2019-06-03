package pages

import (
	. "github.com/go-ozzo/ozzo-validation"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/tokend/regources/generated"
	"net/url"
)

type Params struct {
	Number *string `page:"number"`
	Limit  *string `page:"limit"`
	Offset *string `page:"offset"`
	Cursor *string `page:"cursor"`
	Order  *string `page:"order"`
}

func isNil(i interface{}) error {
	_, isNil := Indirect(i)
	if !isNil {
		return errors.New("must be nil")
	}
	return nil
}

func (p Params) Validate() error {
	errs := Errors{
		"Number": Validate(&p.Number, NilOrNotEmpty),
		"Limit":  Validate(&p.Limit, NilOrNotEmpty),
		"Order":  Validate(&p.Order, NilOrNotEmpty),
	}

	if p.Cursor != nil {
		errs["Cursor"] = Validate(p.Cursor, NotNil)
		errs["Offset"] = Validate(p.Offset, By(isNil))
	}

	if p.Offset != nil {
		errs["Cursor"] = Validate(p.Cursor, By(isNil))
		errs["Offset"] = Validate(p.Offset, NotNil)
	}
	return errs.Filter()
}

func NextPage(links regources.Links) (Params, error) {
	nextRaw := links.Next
	return pageParamsFromQuery(nextRaw)
}


func pageParamsFromQuery(raw string) (Params, error) {
	query, err := url.ParseQuery(raw)
	if err != nil {
		return Params{}, errors.Wrap(err, "failed ot parse query arguments")
	}

	params, err := NewPageParams(query)
	if err != nil {
		return Params{}, errors.Wrap(err, "failed to prepare page params")
	}

	return params, nil
}

func NewPageParams(vals url.Values) (Params, error) {
	params := Params{}
	if number := vals.Get("number"); number != "" {
		params.Number = &number
	}

	if limit := vals.Get("limit"); limit != "" {
		params.Limit = &limit
	}

	if order := vals.Get("order"); order != ""{
		params.Order = &order
	}

	if cursor := vals.Get("cursor"); cursor != "" {
		params.Cursor = &cursor
	}

	if offset := vals.Get("offset"); offset != ""{
		params.Offset = &offset
	}

	return params, params.Validate()
}
