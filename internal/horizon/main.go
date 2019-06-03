package horizon

import (
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/tokend/keypair"
	"net/http"
	"net/url"
)

type Client struct {
	client *http.Client
	log *logan.Entry
	base    *url.URL
	signer keypair.Full
}

func NewClient(base *url.URL) *Client {
	return &Client{
		client: http.DefaultClient,
		base: base,
	}
}

func (c Client) WithSigner(signer keypair.Full) *Client{
	c.signer = signer
	return &c
}
