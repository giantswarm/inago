package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/giantswarm/inago/controller"
)

var (
	stopCmd = &cobra.Command{
		Use:   "stop <group[@slice]...>",
		Short: "Stop a group",
		Long:  "Stop the specified group, or slices",
		Run:   stopRun,
	}
)

func stopRun(cmd *cobra.Command, args []string) {
	newLogger.Debug(newCtx, "cli: starting stop")

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

	if len(newRequestConfig.SliceIDs) == 0 {
		req, err = newController.ExtendWithExistingSliceIDs(req)
		if err != nil {
			newLogger.Error(newCtx, "%#v", maskAny(err))
			os.Exit(1)
		}
	}

	taskObject, err := newController.Stop(newCtx, req)
	if err != nil {
		newLogger.Error(newCtx, "%#v", maskAny(err))
		os.Exit(1)
	}

	maybeBlockWithFeedback(newCtx, blockWithFeedbackCtx{
		Request:    req,
		Descriptor: "stop",
		NoBlock:    globalFlags.NoBlock,
		TaskID:     taskObject.ID,
		Closer:     nil,
	})
}
