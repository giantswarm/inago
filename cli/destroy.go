package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/giantswarm/inago/controller"
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
	newLogger.Debug(newCtx, "cli: starting destroy")

	group := ""
	switch len(args) {
	case 1:
		group = args[0]
	default:
		cmd.Help()
		os.Exit(1)
	}

	newRequestConfig := controller.DefaultRequestConfig()
	newRequestConfig.Group = group
	req := controller.NewRequest(newRequestConfig)

	req, err := newController.ExtendWithExistingSliceIDs(req)
	if err != nil {
		newLogger.Error(newCtx, "%#v", maskAny(err))
		os.Exit(1)
	}

	taskObject, err := newController.Destroy(newCtx, req)
	if err != nil {
		newLogger.Error(newCtx, "%#v", maskAny(err))
		os.Exit(1)
	}

	maybeBlockWithFeedback(newCtx, blockWithFeedbackCtx{
		Request:    req,
		Descriptor: "destroy",
		NoBlock:    globalFlags.NoBlock,
		TaskID:     taskObject.ID,
		Closer:     nil,
	})
}
