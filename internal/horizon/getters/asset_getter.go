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

type AssetPager interface {
	Next() (*regources.AssetListResponse, error)
	Prev() (*regources.AssetListResponse, error)
	Self() (*regources.AssetListResponse, error)
}

type AssetGetter interface {
	SetFilters(filters query.AssetFilters)
	SetIncludes(includes query.AssetIncludes)
	SetPageParams(pageParams page.Params)
	SetParams(params query.AssetParams)

	Filter() query.AssetFilters
	Include() query.AssetIncludes
	Page() page.Params

	ByID(ID string) (*regources.AssetResponse, error)
	List() (*regources.AssetListResponse, error)
}

type AssetHandler interface {
	AssetGetter
	AssetPager
}

type defaultAssetHandler struct {
	base   Getter
	params query.AssetParams

	currentPageLinks *regources.Links
}

func NewDefaultAssetHandler(c *client.Client) *defaultAssetHandler {
	return &defaultAssetHandler{
		base: New(c),
	}
}

func (g *defaultAssetHandler) SetFilters(filters query.AssetFilters) {
	g.params.Filters = filters
}

func (g *defaultAssetHandler) SetIncludes(includes query.AssetIncludes) {
	g.params.Includes = includes
}

func (g *defaultAssetHandler) SetPageParams(pageParams page.Params) {
	g.params.PageParams = pageParams
}

func (g *defaultAssetHandler) SetParams(params query.AssetParams) {
	g.params = params
}

func (g *defaultAssetHandler) Params() query.AssetParams {
	return g.params
}

func (g *defaultAssetHandler) Filter() query.AssetFilters {
	return g.params.Filters
}

func (g *defaultAssetHandler) Include() query.AssetIncludes {
	return g.params.Includes
}

func (g *defaultAssetHandler) Page() page.Params {
	return g.params.PageParams
}

func (g *defaultAssetHandler) ByID(ID string) (*regources.AssetResponse, error) {
	result := &regources.AssetResponse{}
	err := g.base.GetPage(query.AssetByID(ID), g.params, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get record by id", logan.F{
			"id": ID,
		})
	}
	return result, nil
}

func (g *defaultAssetHandler) List() (*regources.AssetListResponse, error) {
	result := &regources.AssetListResponse{}
	err := g.base.GetPage(query.AssetList(), g.params, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get records list", logan.F{
			"query_params": g.params,
		})
	}
	g.currentPageLinks = result.Links
	return result, nil
}

func (g *defaultAssetHandler) Next() (*regources.AssetListResponse, error) {
	if g.currentPageLinks == nil {
		return nil, errors.New("Empty links")
	}
	if g.currentPageLinks.Next == "" {
		return nil, nil
	}
	result := &regources.AssetListResponse{}
	err := g.base.PageFromLink(g.currentPageLinks.Next, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get next page", logan.F{
			"link": g.currentPageLinks.Next,
		})
	}

	return result, nil
}

func (g *defaultAssetHandler) Prev() (*regources.AssetListResponse, error) {
	if g.currentPageLinks == nil {
		return nil, errors.New("Empty links")
	}
	if g.currentPageLinks.Prev == "" {
		return nil, nil
	}

	result := &regources.AssetListResponse{}
	err := g.base.PageFromLink(g.currentPageLinks.Prev, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get next page", logan.F{
			"link": g.currentPageLinks.Next,
		})
	}

	return result, nil
}

func (g *defaultAssetHandler) Self() (*regources.AssetListResponse, error) {
	if g.currentPageLinks == nil {
		return nil, errors.New("Empty links")
	}
	if g.currentPageLinks.Self == "" {
		return nil, nil
	}
	result := &regources.AssetListResponse{}
	err := g.base.PageFromLink(g.currentPageLinks.Self, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get next page", logan.F{
			"link": g.currentPageLinks.Next,
		})
	}

	return result, nil
}
