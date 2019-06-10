// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

package getters

import (
	"github.com/tokend/stellar-deposit-svc/internal/horizon/client"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/page"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/query"
	"gitlab.com/distributed_lab/logan"
	"gitlab.com/distributed_lab/logan/v3/errors"
	regources "gitlab.com/tokend/regources/generated"
)

type TransactionPager interface {
	Next() (*regources.TransactionListResponse, error)
	Prev() (*regources.TransactionListResponse, error)
	Self() (*regources.TransactionListResponse, error)
}

type TransactionGetter interface {
	SetFilters(filters query.TransactionFilters)
	SetIncludes(includes query.TransactionIncludes)
	SetPageParams(pageParams page.Params)
	SetParams(params query.TransactionParams)

	Filter() query.TransactionFilters
	Include() query.TransactionIncludes
	Page() page.Params

	ByID(ID string) (*regources.TransactionResponse, error)
	List() (*regources.TransactionListResponse, error)
}

type TransactionHandler interface {
	TransactionGetter
	TransactionPager
}

type defaultTransactionHandler struct {
	base   Getter
	params query.TransactionParams

	currentPageLinks *regources.Links
}

func NewDefaultTransactionHandler(c *client.Client) *defaultTransactionHandler {
	return &defaultTransactionHandler{
		base: New(c),
	}
}

func (g *defaultTransactionHandler) SetFilters(filters query.TransactionFilters) {
	g.params.Filters = filters
}

func (g *defaultTransactionHandler) SetIncludes(includes query.TransactionIncludes) {
	g.params.Includes = includes
}

func (g *defaultTransactionHandler) SetPageParams(pageParams page.Params) {
	g.params.PageParams = pageParams
}

func (g *defaultTransactionHandler) SetParams(params query.TransactionParams) {
	g.params = params
}

func (g *defaultTransactionHandler) Params() query.TransactionParams {
	return g.params
}

func (g *defaultTransactionHandler) Filter() query.TransactionFilters {
	return g.params.Filters
}

func (g *defaultTransactionHandler) Include() query.TransactionIncludes {
	return g.params.Includes
}

func (g *defaultTransactionHandler) Page() page.Params {
	return g.params.PageParams
}

func (g *defaultTransactionHandler) ByID(ID string) (*regources.TransactionResponse, error) {
	result := &regources.TransactionResponse{}
	err := g.base.GetPage(query.TransactionByID(ID), g.params, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get record by id", logan.F{
			"id": ID,
		})
	}
	return result, nil
}

func (g *defaultTransactionHandler) List() (*regources.TransactionListResponse, error) {
	result := &regources.TransactionListResponse{}
	err := g.base.GetPage(query.TransactionList(), g.params, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get records list", logan.F{
			"query_params": g.params,
		})
	}
	g.currentPageLinks = result.Links
	return result, nil
}

func (g *defaultTransactionHandler) Next() (*regources.TransactionListResponse, error) {
	if g.currentPageLinks == nil {
		return nil, errors.New("Empty links")
	}
	if g.currentPageLinks.Next == "" {
		return nil, nil
	}
	result := &regources.TransactionListResponse{}
	err := g.base.PageFromLink(g.currentPageLinks.Next, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get next page", logan.F{
			"link": g.currentPageLinks.Next,
		})
	}

	return result, nil
}

func (g *defaultTransactionHandler) Prev() (*regources.TransactionListResponse, error) {
	if g.currentPageLinks == nil {
		return nil, errors.New("Empty links")
	}
	if g.currentPageLinks.Prev == "" {
		return nil, nil
	}

	result := &regources.TransactionListResponse{}
	err := g.base.PageFromLink(g.currentPageLinks.Prev, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get next page", logan.F{
			"link": g.currentPageLinks.Next,
		})
	}

	return result, nil
}

func (g *defaultTransactionHandler) Self() (*regources.TransactionListResponse, error) {
	if g.currentPageLinks == nil {
		return nil, errors.New("Empty links")
	}
	if g.currentPageLinks.Self == "" {
		return nil, nil
	}
	result := &regources.TransactionListResponse{}
	err := g.base.PageFromLink(g.currentPageLinks.Self, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get next page", logan.F{
			"link": g.currentPageLinks.Next,
		})
	}

	return result, nil
}
