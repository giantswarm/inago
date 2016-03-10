package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/giantswarm/inago/logging"
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
	logger := logging.GetLogger()

	req, err := createRequest(args)
	if err != nil {
		logger.Error(nil, "%#v", maskAny(err))
		os.Exit(1)
	}

	taskObject, err := newController.Stop(req)
	if err != nil {
		logger.Error(nil, "%#v", maskAny(err))
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
