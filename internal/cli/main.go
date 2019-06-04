package cli

import (
	"context"
	"fmt"
	"github.com/urfave/cli"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"github.com/tokend/stellar-deposit-svc/internal/config"
	"github.com/tokend/stellar-deposit-svc/internal/gather"
)

func Run(args []string) bool {
	log := logan.New()
	var cfg config.Config
	defer func() {
		if rvr := recover(); rvr != nil {
			log.WithRecover("app panicked")
		}
	}()

	app := cli.NewApp()

	initialize := func (_ *cli.Context) error {
		getter, err := kv.FromEnv()
		if err != nil {
			if err == kv.ErrNoBackends {
				fmt.Println("Could not get config - is KV_VIPER_FILE env var set?")
			}
			return errors.Wrap(err, "failed to get config")
		}
		cfg = config.NewConfig(getter)

		return nil
	}

	app.Commands = cli.Commands{
		{
			Name: "run",
			Subcommands: cli.Commands{
				{
					Name: "deposit",
					Before: initialize,
					Action: func(_ *cli.Context) error {
						service := gather.NewService(gather.Opts{
							Log: log,
							Config: cfg,
						})

						service.Run(context.Background())
						return errors.New("service died")
					},
				},
			},

		},
	}
	if err := app.Run(args); err != nil {
		log.WithError(err)
		return false
	}
	return true
}
