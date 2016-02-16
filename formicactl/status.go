package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/giantswarm/formica/fleet"
)

var (
	statusCmd = &cobra.Command{
		Use:   "status [group]",
		Short: "status of a group",
		Long:  "status of a group",
		Run:   statusRun,
	}
)

func statusRun(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Help()
		os.Exit(1)
	}

	status, err := newController.GetStatus(args[0])
	if err != nil {
		fmt.Printf("%#v\n", err)
		os.Exit(1)
	}

	err = fleet.PrintStatus(status)
	if err != nil {
		fmt.Printf("%#v\n", err)
		os.Exit(1)
	}
}
