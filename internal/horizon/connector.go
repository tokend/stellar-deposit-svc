package horizon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/tokend/go/xdrbuild"
	"gitlab.com/tokend/regources/generated"
	"gitlab.com/tokend/stellar-deposit-svc/internal/horizon/assets"
	"gitlab.com/tokend/stellar-deposit-svc/internal/horizon/pages"
	"gitlab.com/tokend/stellar-deposit-svc/internal/horizon/transactions"
	"net/http"
)

var (
	ErrLinksMissing = errors.New("links missing")
)

const (
	streamPageLimit = 100
)

func (c *Client) PageFromLink(link string, v interface{}) error {
	u, err := c.urlFromLink(link)
	if err != nil {
		return errors.Wrap(err, "failed to get url from link")
	}
	r, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return errors.Wrap(err, "failed to prepare request")
	}

	resp, err := c.performRequest(r)
	if err != nil {
		return errors.Wrap(err, "failed to perform request")
	}

	response := bytes.NewReader(resp)
	decoder := json.NewDecoder(response)

	err = decoder.Decode(v)
	if err != nil {
		return errors.Wrap(err, "failed to parse response")
	}

	return nil
}

func (c *Client) NextPage(links *regources.Links, v interface{}) error {
	if links == nil {
		return ErrLinksMissing
	}
	return c.PageFromLink(links.Next, v)
}

func HasMorePages(links *regources.Links) bool {
	if links == nil {
		return false
	}
	return links.Next != ""
}

func (c *Client) GetList(endpoint string, params QueryParams, result interface{}) error {
	request, err := c.NewRequest("GET", endpoint, params, nil)
	if err != nil {
		return errors.Wrap(err, "failed to prepare request")
	}

	resp, err := c.performRequest(request)
	if err != nil {
		return errors.Wrap(err, "failed to perform request")
	}

	response := bytes.NewReader(resp)
	decoder := json.NewDecoder(response)
	err = decoder.Decode(result)
	if err != nil {
		return errors.Wrap(err, "failed to parse response")
	}

	return nil
}

func (c *Client) GetSingle(endpoint string, params QueryParams, result interface{}) error {
	request, err := c.NewRequest("GET", endpoint, params, nil)
	if err != nil {
		return errors.Wrap(err, "failed to prepare request")
	}

	resp, err := c.performRequest(request)
	if err != nil {
		return errors.Wrap(err, "failed to perform request")
	}
	err = decodeResponse(resp, &result)
	if err != nil {
		return errors.Wrap(err, "failed to decode response")
	}
	return nil
}

func (c *Client) AssetByCode(code string, params QueryParams) (*regources.AssetResponse, error) {

	result := &regources.AssetResponse{}
	err := c.GetSingle(assets.ByID(code), params, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get asset by code", logan.F{
			"asset_code": code,
		})
	}

	return result, nil
}

func (c *Client) TransactionByID(id string, params QueryParams) (*regources.TransactionResponse, error) {

	result := &regources.TransactionResponse{}
	err := c.GetSingle(transactions.ByID(id), params, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get tx by id", logan.F{
			"tx_id": id,
		})
	}

	return result, nil
}

func (c *Client) Assets(params QueryParams) (*regources.AssetListResponse, error) {

	result := &regources.AssetListResponse{}
	err := c.GetList(assets.List(), params, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get asset list", logan.F{
			"query_params": params,
		})
	}

	return result, nil
}

func (c *Client) Transactions(params QueryParams) (*regources.TransactionListResponse, error) {

	result := &regources.TransactionListResponse{}
	err := c.GetList(transactions.List(), params, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get transaction list", logan.F{
			"query_params": params,
		})
	}

	return result, nil
}

func (c *Client) StreamTransactions(ctx context.Context, changeTypes, entryTypes []int,
) (<-chan regources.TransactionResponse, <-chan error) {
	txChan := make(chan regources.TransactionResponse)
	errChan := make(chan error)
	defer close(txChan)
	defer close(errChan)
	limit := fmt.Sprintf("%d", streamPageLimit)
	params := transactions.Params{
		Includes: transactions.Includes{
			LedgerEntryChanges: true,
		},
		Filters: transactions.Filters{
			ChangeTypes: changeTypes,
			EntryTypes:  entryTypes,
		},
		PageParams: pages.Params{
			Limit: &limit,
		},
	}
	txPage, err := c.Transactions(params)

	processedOnPage := make(map[string]bool)
	go func() {
		for {
			if err != nil {
				errChan <- err
				return
			}
			tx := regources.TransactionResponse{}
			for _, transaction := range txPage.Data {
				if _, ok := processedOnPage[transaction.ID]; ok {
					continue
				}
				processedOnPage[transaction.ID] = true

				tx.Data = transaction
				tx.Meta = txPage.Meta

				for _, relation := range transaction.Relationships.LedgerEntryChanges.Data {
					tx.Included.Add(txPage.Included.MustLedgerEntryChange(relation))
				}

				txChan <- tx
			}

			if len(txPage.Data) < streamPageLimit {
				err = c.PageFromLink(txPage.Links.Self, txPage)
			} else {
				err = c.PageFromLink(txPage.Links.Next, txPage)
				processedOnPage = make(map[string]bool)
			}
		}
	}()

	return txChan, errChan

}

func (c *Client) State() (*regources.HorizonStateResponse, error) {
	result := &regources.HorizonStateResponse{}
	err := c.GetSingle("/v3/", nil, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get horizon state")
	}

	return result, nil
}

func (c *Client) Builder() (*xdrbuild.Builder, error) {
	state, err := c.State()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get network passphrase and tx expiration period")
	}

	return xdrbuild.NewBuilder(state.Data.Attributes.NetworkPassphrase, state.Data.Attributes.TxExpirationPeriod), nil
}
