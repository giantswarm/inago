package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	destroyCmd = &cobra.Command{
		Use:   "destroy [group [group..]]",
		Short: "Destroys the specified group or slices",
		Long:  "Destroys a group or the specified slices",
		Run:   destroyRun,
	}
)

func destroyRun(cmd *cobra.Command, args []string) {
	req, err := createRequest(args)
	if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}

	closer := make(<-chan struct{})
	taskObject, err := newController.Destroy(req, closer)
	if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}

	maybeBlockWithFeedback(blockWithFeedbackCtx{
		Request:    req,
		Descriptor: "destroy",
		NoBlock:    globalFlags.NoBlock,
		TaskID:     taskObject.ID,
		Closer:     nil,
	})
}
