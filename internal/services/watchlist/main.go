package watchlist

import (
	"github.com/tokend/stellar-deposit-svc/internal/horizon/getters"
	"gitlab.com/distributed_lab/logan/v3"
)

type Service struct {
	streamer  getters.AssetHandler
	log       *logan.Entry
	watchlist map[string]bool
	toAdd     chan Details
	toRemove  chan string
}

type Opts struct {
	Streamer   getters.AssetHandler
	Log        *logan.Entry
}

func New(opts Opts) *Service {
	toAdd := make(chan Details)
	toRemove := make(chan string)
	return &Service{
		streamer:  opts.Streamer,
		log:       opts.Log.WithField("service", "watchlist"),
		watchlist: make(map[string]bool),
		toRemove:  toRemove,
		toAdd:     toAdd,
	}
}
