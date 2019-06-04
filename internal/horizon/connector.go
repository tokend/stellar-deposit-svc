package horizon

import (
	"bytes"
	"encoding/json"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/tokend/go/xdrbuild"
	"gitlab.com/tokend/keypair"
	"gitlab.com/tokend/regources/generated"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/client"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/query"
	"net/http"
	"net/url"
)

var (
	ErrLinksMissing = errors.New("links missing")
)

type Interface interface {
	GetList(endpoint string, params query.Params, result interface{}) error
	GetSingle(endpoint string, params query.Params, result interface{}) error
	PageFromLink(link string, v interface{}) error
}

type Connector struct {
	*client.Client
}

func NewConnector(base *url.URL) *Connector {
	cl := client.New(http.DefaultClient, base)
	return &Connector{
		Client: cl,
	}
}

func (c *Connector) WithSigner (signer keypair.Full) *Connector {
	return &Connector{
		c.Client.WithSigner(signer),
	}
}
func (c *Connector) PageFromLink(link string, v interface{}) error {
	resp, err := c.Get(link)
	if err != nil {
		return errors.Wrap(err, "failed to get page")
	}

	response := bytes.NewReader(resp)
	decoder := json.NewDecoder(response)

	err = decoder.Decode(v)
	if err != nil {
		return errors.Wrap(err, "failed to parse response")
	}

	return nil
}

func (c *Connector) NextPage(links *regources.Links, v interface{}) error {
	if links == nil {
		return ErrLinksMissing
	}
	return c.PageFromLink(links.Next, v)
}
func (c *Connector) GetList(endpoint string, params query.Params, result interface{}) error {
	q, err := query.Prepare(params)
	if err != nil {
		return errors.Wrap(err, "failed to prepare query")
	}
	uri, err  := c.WithQuery(endpoint, q)
	if err != nil {
		return errors.Wrap(err, "failed to resolve request URI", logan.F{
			"endpoint": endpoint,
			"query": params,
		})
	}

	resp, err := c.Get(uri)
	if err != nil {
		return errors.Wrap(err, "failed to perform get")
	}

	response := bytes.NewReader(resp)
	decoder := json.NewDecoder(response)
	err = decoder.Decode(result)
	if err != nil {
		return errors.Wrap(err, "failed to parse response")
	}

	return nil
}

func (c *Connector) GetSingle(endpoint string, params query.Params, result interface{}) error {
	q, err := query.Prepare(params)
	if err != nil {
		return errors.Wrap(err, "failed to prepare query")
	}
	uri, err  := c.WithQuery(endpoint, q)
	if err != nil {
		return errors.Wrap(err, "failed to resolve request URI", logan.F{
			"endpoint": endpoint,
			"query": params,
		})
	}
	resp, err := c.Get(uri)
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

func (c *Connector) State() (*regources.HorizonStateResponse, error) {
	result := &regources.HorizonStateResponse{}
	err := c.GetSingle("/v3/", nil, result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get horizon state")
	}

	return result, nil
}

func (c *Connector) Builder() (*xdrbuild.Builder, error) {
	state, err := c.State()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get network passphrase and tx expiration period")
	}

	return xdrbuild.NewBuilder(state.Data.Attributes.NetworkPassphrase, state.Data.Attributes.TxExpirationPeriod), nil
}
