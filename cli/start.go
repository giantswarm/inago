package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	startCmd = &cobra.Command{
		Use:   "start [group [group..]]",
		Short: "Starts the specified group or slices",
		Long:  "Starts a group or the specified slices",
		Run:   startRun,
	}
)

func startRun(cmd *cobra.Command, args []string) {
	req, err := createRequest(args)
	if err != nil {
		newLogger.Error(nil, "%#v", maskAny(err))
		os.Exit(1)
	}

	taskObject, err := newController.Start(req)
	if err != nil {
		newLogger.Error(nil, "%#v", maskAny(err))
		os.Exit(1)
	}

	maybeBlockWithFeedback(blockWithFeedbackCtx{
		Request:    req,
		Descriptor: "start",
		NoBlock:    globalFlags.NoBlock,
		TaskID:     taskObject.ID,
		Closer:     nil,
	})
}
