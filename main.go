package main

import (
	"os"

	"github.com/giantswarm/inago/cli"
	"github.com/giantswarm/inago/logging"
)

func main() {
	if err := cli.MainCmd.Execute(); err != nil {
		logging.GetLogger().Error(nil, err.Error())
		os.Exit(-1)
	}
}
