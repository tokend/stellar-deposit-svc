package horizon

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/google/jsonapi"
	regources "gitlab.com/tokend/regources/generated"
	"net/http"

	"gitlab.com/distributed_lab/logan/v3/errors"
)

var (
	ErrSubmitTimeout              = errors.New("submit timed out")
	ErrSubmitInternal             = errors.New("internal submit error")
	ErrSubmitUnexpectedStatusCode = errors.New("Unexpected unsuccessful status code.")
)

type Error interface {
	error
	Status() int
	Body() []byte
	Path() string
}

func (c *Client) Submit(ctx context.Context, envelope string) (*regources.TransactionResponse, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(&regources.SubmitTransactionBody{
		Tx: envelope,
	})
	if err != nil {
		panic(errors.Wrap(err, "failed to marshal request"))
	}
	req, err := c.NewRequest("POST", "/v3/transactions", nil, &buf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare request")
	}


	response, err := c.performRequest(req)
	if err == nil {
		var success regources.TransactionResponse
		if err := json.Unmarshal(response, &success); err != nil {
		}
		return &success, nil
	}

	cerr := errors.Cause(err).(Error)

	// go through known response codes and try to build meaningful result
	switch cerr.Status() {
	case http.StatusGatewayTimeout: // timeout
		return nil, ErrSubmitTimeout
	case http.StatusBadRequest: // rejected or malformed
		// check which error it was exactly, might be useful for consumer
		var errorObj jsonapi.ErrorObject
		if err := json.Unmarshal(response, &errorObj); err != nil {
			panic(errors.Wrap(err, "failed to unmarshal horizon response"))
		}
		return nil, &errorObj
	case http.StatusInternalServerError: // internal error
		return nil, ErrSubmitInternal
	default:
		panic("unexpected submission result")
	}
}