package cli

import (
	"fmt"
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
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}

	closer := make(chan struct{}, 1)
	taskObject, err := newController.Start(req, closer)
	if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
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
