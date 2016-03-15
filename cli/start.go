package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/giantswarm/inago/controller"
)

var (
	startCmd = &cobra.Command{
		Use:   "start <group|slice...>",
		Short: "Start a group",
		Long:  "Start the specified group, or slices",
		Run:   startRun,
	}
)

func startRun(cmd *cobra.Command, args []string) {
	newLogger.Debug(newCtx, "cli: starting start")

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

	taskObject, err := newController.Start(newCtx, req)
	if err != nil {
		newLogger.Error(newCtx, "%#v", maskAny(err))
		os.Exit(1)
	}

	maybeBlockWithFeedback(newCtx, blockWithFeedbackCtx{
		Request:    req,
		Descriptor: "start",
		NoBlock:    globalFlags.NoBlock,
		TaskID:     taskObject.ID,
		Closer:     nil,
	})
}
