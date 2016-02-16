package main

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	createCmd = &cobra.Command{
		Use:   "create [group]",
		Short: "create a group",
		Long:  "create a group",
		Run:   createRun,
	}
)

func createRun(cmd *cobra.Command, args []string) {
	if len(args) < 1 || len(args) > 1 {
		cmd.Help()
		os.Exit(1)
	}
}
