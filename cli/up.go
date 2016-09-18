package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	upCmd = &cobra.Command{
		Use:   "up <group> [scale]",
		Short: "Bring a group up",
		Long:  "Submit a group, with an optional scale, and start it",
		Run:   upRun,
	}
)

func upRun(cmd *cobra.Command, args []string) {
	group, err := base(args[0])
	if err != nil {
		newLogger.Error(newCtx, "%#v\n", maskAny(err))
		os.Exit(1)
	}
	submitRun(cmd, args)
	startRun(cmd, []string{group})
}
