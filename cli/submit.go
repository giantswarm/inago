package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	submitCmd = &cobra.Command{
		Use:   "submit [group]",
		Short: "submit a group",
		Long:  "submit a group",
		Run:   submitRun,
	}
)

func submitRun(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Help()
		os.Exit(1)
	}
}
