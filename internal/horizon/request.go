package horizon

import (
	"bytes"
	"encoding/json"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	depkeypair "gitlab.com/tokend/go/keypair"
	"gitlab.com/tokend/go/signcontrol"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

func (c *Client) NewRequest(method string, endpoint string, qp QueryParams, body io.Reader) (*http.Request, error) {
	query, err := prepareQuery(qp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare query params")
	}
	u, err := c.resolveURL(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve url")
	}
	r, err := http.NewRequest(method, u, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare request")
	}
	r.URL.RawQuery = query.Encode()

	return r, nil
}

func (c *Client) urlFromLink(link string) (string, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", errors.New("failed to parse link")
	}

	return c.base.ResolveReference(u).String(), nil
}

func (c *Client) resolveURL(endpoint string) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse endpoint into URL")
	}

	return c.base.ResolveReference(u).String(), nil
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

func decodeResponse(bb []byte, v interface{}) error {
	response := bytes.NewReader(bb)
	decoder := json.NewDecoder(response)

	err := decoder.Decode(v)
	if err != nil {
		return errors.Wrap(err, "failed to parse response")
	}
	return nil
}
