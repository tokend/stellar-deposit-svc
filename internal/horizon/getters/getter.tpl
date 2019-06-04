package getters


import (
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	regources "gitlab.com/tokend/regources/generated"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/connector"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/query"
)


type TemplateGetter struct {
	horizon.Interface
}



func (g *TemplateGetter) TemplateByID(ID string, params query.TemplateParams) (*regources.TemplateResponse, error) {
	result := &regources.TemplateResponse{}
	err := g.GetSingle(query.TemplateByID(ID), params, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get record by id", logan.F{
			"id": ID,
		})
	}

	return result, nil
}

func (g *TemplateGetter) TemplateList(params query.Params) (*regources.TemplateListResponse, error) {
	result := &regources.TemplateListResponse{}
	err := g.GetList(query.TemplateList(), params, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get records list", logan.F{
			"query_params": params,
		})
	}

	return result, nil
}

func (g *TemplateGetter) Next(links *regources.Links) (*regources.TemplateListResponse, error){
	if links == nil{
		return nil, errors.New("Empty links")
	}
	result := &regources.TemplateListResponse{}
	err := g.PageFromLink(links.Next, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get next page", logan.F{
			"link": links.Next,
		})
	}

	return result, nil
}
