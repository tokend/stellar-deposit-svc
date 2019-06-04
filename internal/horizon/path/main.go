package path

import (
	"gitlab.com/distributed_lab/logan/v3/errors"
	"net/url"
)

type Resolver interface {
	URL (string) (string, error)
	WithQuery(string, url.Values) (string, error)
}

type resolver struct {
	base *url.URL
}

func (r resolver) URL(endpoint string) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse endpoint into URL")
	}

	return r.base.ResolveReference(u).String(), nil
}

func (r resolver) WithQuery(endpoint string, values url.Values) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse endpoint into URL")
	}

	resolved := r.base.ResolveReference(u)
	resolved.RawQuery = values.Encode()

	return resolved.String(), nil
}

func NewResolver(base *url.URL) Resolver {
	return &resolver{
		base: base,
	}
}

