package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	stopCmd = &cobra.Command{
		Use:   "stop [group [group..]]",
		Short: "Stops the specified group or slices",
		Long:  "Stops a group or the specified slices",
		Run:   stopRun,
	}
)

func stopRun(cmd *cobra.Command, args []string) {
	req, err := createRequest(args)
	if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}

	taskObject, err := newController.Stop(req)
	if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}

	maybeBlockWithFeedback(blockWithFeedbackCtx{
		Request:    req,
		Descriptor: "stop",
		NoBlock:    globalFlags.NoBlock,
		TaskID:     taskObject.ID,
		Closer:     nil,
	})
}
