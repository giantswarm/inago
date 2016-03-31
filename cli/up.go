package cli

import (
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
	submitRun(cmd, args)

	// If a scale argument has been passed to submit,
	// remove it from the args list, as start doesn't want it.
	if len(args) > 1 {
		args = args[:1]
	}

	startRun(cmd, args)
}
