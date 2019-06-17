package watchlist

import (
	"github.com/tokend/stellar-deposit-svc/internal/horizon/getters"
	"gitlab.com/distributed_lab/logan/v3"
)

type Service struct {
	streamer getters.AssetHandler
	log      *logan.Entry
	owner    string
	ch       chan Details
}

type Opts struct {
	Streamer   getters.AssetHandler
	Log        *logan.Entry
	AssetOwner string
}

func New(opts Opts) *Service {
	ch := make(chan Details)
	return &Service{
		streamer: opts.Streamer,
		owner:    opts.AssetOwner,
		log:      opts.Log.WithField("service", "watchlist"),
		ch:       ch,
	}
}
