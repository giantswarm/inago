package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/giantswarm/inago/logging"
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
	logger := logging.GetLogger()

	req, err := createRequest(args)
	if err != nil {
		logger.Error(nil, "%#v", maskAny(err))
		os.Exit(1)
	}

	taskObject, err := newController.Destroy(req)
	if err != nil {
		logger.Error(nil, "%#v", maskAny(err))
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
