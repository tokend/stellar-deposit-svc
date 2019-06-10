package submit

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/tokend/stellar-deposit-svc/internal/horizon/client"
	"net/http"

	"github.com/google/jsonapi"
	regources "gitlab.com/tokend/regources/generated"

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

type Interface interface {
	Submit(ctx context.Context, envelope string) (*regources.TransactionResponse, error)
}

type submitter struct {
	*client.Client
}

func New(cl *client.Client) *submitter {
	return &submitter{
		Client: cl,
	}
}

func (s *submitter) Submit(ctx context.Context, envelope string) (*regources.TransactionResponse, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(&regources.SubmitTransactionBody{
		Tx: envelope,
	})
	if err != nil {
		panic(errors.Wrap(err, "failed to marshal request"))
	}
	response, err := s.Post("/v3/transactions", &buf)
	if err == nil {
		var success regources.TransactionResponse
		if err := json.Unmarshal(response, &success); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal transaction response")
		}
		return &success, nil
	}

	cerr, ok := errors.Cause(err).(Error)
	if !ok {
		return nil, err
	}
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
		return nil, ErrSubmitUnexpectedStatusCode
	}
}
