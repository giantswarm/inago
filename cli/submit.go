package cli

import (
	"fmt"
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
	req, err := createRequestWithContent(args)
	if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}

	closer := make(chan struct{}, 1)
	taskObject, err := newController.Submit(req, closer)
	if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}

	maybeBlockWithFeedback(blockWithFeedbackCtx{
		Request:    req,
		Descriptor: "submit",
		NoBlock:    globalFlags.NoBlock,
		TaskID:     taskObject.ID,
		Closer:     nil,
	})
}
