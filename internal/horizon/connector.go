package horizon

import (
	"bytes"
	"encoding/json"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/client"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/submit"
	"net/http"
	"net/url"


	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/tokend/go/xdrbuild"
	"gitlab.com/tokend/keypair"
	"gitlab.com/tokend/regources/generated"
)

type Connector interface {
	State() (*regources.HorizonStateResponse, error)
	Builder() (*xdrbuild.Builder, error)
	Submitter() (submit.Interface, error)
}

type connector struct {
	*client.Client
}

func NewConnector(base *url.URL) *connector {
	cl := client.New(http.DefaultClient, base)
	return &connector{
		Client: cl,
	}
}

func (c *connector) WithSigner(signer keypair.Full) *connector {
	return &connector{
		c.Client.WithSigner(signer),
	}
}

func (c *connector) State() (*regources.HorizonStateResponse, error) {
	result := &regources.HorizonStateResponse{}
	respBB, err := c.Get("/v3/")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get horizon state")
	}
	buf := bytes.NewBuffer(respBB)
	decoder := json.NewDecoder(buf)
	if err := decoder.Decode(result); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal horizon response")
	}

	return result, nil
}

func (c *connector) Builder() (*xdrbuild.Builder, error) {
	state, err := c.State()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get network passphrase and tx expiration period")
	}

	return xdrbuild.NewBuilder(state.Data.Attributes.NetworkPassphrase, state.Data.Attributes.TxExpirationPeriod), nil
}

func (c *connector) Submitter() (submit.Interface, error) {
	return submit.New(c.Client), nil
}
