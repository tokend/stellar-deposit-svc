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

type ResourcePager interface {
	Next() (*regources.TransactionListResponse, error)
	Prev() (*regources.TransactionListResponse, error)
	Self() (*regources.TransactionListResponse, error)
	First() (*regources.TransactionListResponse, error)
}

type ResourceGetter interface {
	SetFilters(filters query.ResourceFilters)
	SetIncludes(includes query.ResourceIncludes)
	SetPageParams(pageParams page.Params)
	SetParams(params query.ResourceParams)

	Filter() query.ResourceFilters
	Include() query.ResourceIncludes
	Page() page.Params

	ByID(ID string) (*regources.TransactionResponse, error)
	List() (*regources.TransactionListResponse, error)
}

type ResourceHandler interface {
	ResourceGetter
	ResourcePager
}

type defaultResourceHandler struct {
	base   Getter
	params query.ResourceParams

	currentPageLinks *regources.Links
}

func NewDefaultResourceHandler(c *client.Client) *defaultResourceHandler {
	return &defaultResourceHandler{
		base: New(c),
	}
}

func (g *defaultResourceHandler) SetFilters(filters query.ResourceFilters) {
	g.params.Filters = filters
}

func (g *defaultResourceHandler) SetIncludes(includes query.ResourceIncludes) {
	g.params.Includes = includes
}

func (g *defaultResourceHandler) SetPageParams(pageParams page.Params) {
	g.params.PageParams = pageParams
}

func (g *defaultResourceHandler) SetParams(params query.ResourceParams) {
	g.params = params
}

func (g *defaultResourceHandler) Params() query.ResourceParams {
	return g.params
}

func (g *defaultResourceHandler) Filter() query.ResourceFilters {
	return g.params.Filters
}

func (g *defaultResourceHandler) Include() query.ResourceIncludes {
	return g.params.Includes
}

func (g *defaultResourceHandler) Page() page.Params {
	return g.params.PageParams
}

func (g *defaultResourceHandler) ByID(ID string) (*regources.TransactionResponse, error) {
	result := &regources.TransactionResponse{}
	err := g.base.GetPage(query.ResourceByID(ID), g.params.Includes, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get record by id", logan.F{
			"id": ID,
		})
	}
	return result, nil
}

func (g *defaultResourceHandler) List() (*regources.TransactionListResponse, error) {
	result := &regources.TransactionListResponse{}
	err := g.base.GetPage(query.ResourceList(), g.params, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get records list", logan.F{
			"query_params": g.params,
		})
	}
	g.currentPageLinks = result.Links
	return result, nil
}

func (g *defaultResourceHandler) Next() (*regources.TransactionListResponse, error) {
	if g.currentPageLinks == nil {
		return nil, errors.New("Empty links")
	}
	if g.currentPageLinks.Next == "" {
		return nil, errors.From(errors.New("No link to page"), logan.F{
			"page": "next",
		})
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

func (g *defaultResourceHandler) Prev() (*regources.TransactionListResponse, error) {
	if g.currentPageLinks == nil {
		return nil, errors.New("Empty links")
	}
	if g.currentPageLinks.Prev == "" {
		return nil, errors.From(errors.New("No link to page"), logan.F{
			"page": "prev",
		})
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

func (g *defaultResourceHandler) Self() (*regources.TransactionListResponse, error) {
	if g.currentPageLinks == nil {
		return nil, errors.New("Empty links")
	}
	if g.currentPageLinks.Self == "" {
		return nil, errors.From(errors.New("No link to page"), logan.F{
			"page": "self",
		})
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

func (g *defaultResourceHandler) First() (*regources.TransactionListResponse, error) {
	if g.currentPageLinks == nil {
		return nil, errors.New("Empty links")
	}
	if g.currentPageLinks.First == "" {
		return nil, errors.From(errors.New("No link to page"), logan.F{
			"page": "first",
		})
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
