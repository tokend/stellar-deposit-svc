package main

import (
	"github.com/tokend/stellar-deposit-svc/internal/cli"
	"os"
)

func main() {
	if !cli.Run(os.Args) {
		os.Exit(1)
	}
}
