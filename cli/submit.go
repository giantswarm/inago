package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/giantswarm/inago/validator"
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

	ok, err := validator.ValidateRequest(req)
	if !ok {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}

	taskObject, err := newController.Submit(req)
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
