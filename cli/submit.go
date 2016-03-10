package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/giantswarm/inago/logging"
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
	logger := logging.GetLogger()

	req, err := createRequestWithContent(args)
	if err != nil {
		logger.Error(nil, "%#v", maskAny(err))
		os.Exit(1)
	}

	taskObject, err := newController.Submit(req)
	if err != nil {
		logger.Error(nil, "%#v", maskAny(err))
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
