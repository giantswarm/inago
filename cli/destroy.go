package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/giantswarm/inago/controller"
)

var (
	destroyCmd = &cobra.Command{
		Use:   "destroy <group[@slice]...>",
		Short: "Destroy a group",
		Long:  "Destroy the specified group, or slices",
		Run:   destroyRun,
	}
)

func destroyRun(cmd *cobra.Command, args []string) {
	newLogger.Debug(newCtx, "cli: starting destroy")

	if len(args) == 0 {
		cmd.Help()
		os.Exit(1)
	}

	var err error
	newRequestConfig := controller.DefaultRequestConfig()
	newRequestConfig.Group, newRequestConfig.SliceIDs, err = parseGroupCLIArgs(args)
	if err != nil {
		newLogger.Error(newCtx, "%#v", maskAny(err))
		os.Exit(1)
	}
	req := controller.NewRequest(newRequestConfig)

	// in case no slice id was provided, we extend the request with all
	// slice ids seen in fleet
	if len(newRequestConfig.SliceIDs) == 0 {
		req, err = newController.ExtendWithExistingSliceIDs(req)
		if err != nil {
			newLogger.Error(newCtx, "%#v", maskAny(err))
			os.Exit(1)
		}
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
