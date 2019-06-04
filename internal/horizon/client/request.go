package client

import (
	"github.com/tokend/stellar-deposit-svc/internal/horizon/query"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	depkeypair "gitlab.com/tokend/go/keypair"
	"gitlab.com/tokend/go/signcontrol"
	"io"
	"io/ioutil"
	"net/http"
)

func (c *Client) Get(endpoint string) ([]byte, error) {
	u, err := c.URL(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve url")
	}
	r, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare request")
	}

	return c.performRequest(r)
}

func (c *Client) Put(endpoint string, body io.Reader) ([]byte, error) {
	u, err := c.URL(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve url")
	}
	r, err := http.NewRequest("PUT", u, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare request")
	}

	return c.performRequest(r)
}

func (c *Client) Post(endpoint string, body io.Reader) ([]byte, error) {
	u, err := c.URL(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve url")
	}
	r, err := http.NewRequest("POST", u, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare request")
	}

	return c.performRequest(r)
}

func (c *Client) newRequest(method string, endpoint string, qp query.Params, body io.Reader) (*http.Request, error) {
	q, err := query.Prepare(qp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare query params")
	}
	u, err := c.URL(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve url")
	}
	r, err := http.NewRequest(method, u, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare request")
	}
	r.URL.RawQuery = q.Encode()

	return r, nil
}
func (c *Client) do(r *http.Request) (int, []byte, error) {

	// ensure content-type just in case
	r.Header.Set("content-type", "application/json")

	if c.signer != nil {
		err := signcontrol.SignRequest(r, depkeypair.MustParse(c.signer.Seed()))
		if err != nil {
			return 0, nil, errors.Wrap(err, "failed to sign request")
		}
	}

	response, err := c.client.Do(r)
	if err != nil {
		return 0, nil, errors.Wrap(err, "failed to perform http request")
	}

	defer response.Body.Close()

	respBB, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, nil, errors.Wrap(err, "Failed to read response body", logan.F{
			"status_code": response.StatusCode,
		})
	}

	return response.StatusCode, respBB, nil
}

func (c *Client) performRequest(r *http.Request) ([]byte, error) {
	code, resp, err := c.do(r)
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}

	return handleResponse(resp, code)
}

func handleResponse(resp []byte, code int) ([]byte, error) {
	if isStatusCodeSuccessful(code) {
		return resp, nil
	}

	return resp, errors.New(http.StatusText(code))
}

func isStatusCodeSuccessful(code int) bool {
	return code >= 200 && code < 300
}