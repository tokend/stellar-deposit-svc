package client

import (
	"github.com/tokend/stellar-deposit-svc/internal/horizon/path"
	"gitlab.com/tokend/keypair"
	"net/http"
	"net/url"
)

type Client struct {
	client *http.Client
	signer keypair.Full
	path.Resolver
}

func New(client *http.Client, base *url.URL) *Client {
	return &Client{
		client: client,
		Resolver: path.NewResolver(base),
	}
}

func (c *Client) WithSigner(signer keypair.Full) *Client{
	return &Client{
		client: c.client,
		signer: signer,
		Resolver: c.Resolver,
	}
}
