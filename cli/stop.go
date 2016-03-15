package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/giantswarm/inago/controller"
)

var (
	stopCmd = &cobra.Command{
		Use:   "stop <group|slice...>",
		Short: "Stop a group",
		Long:  "Stop the specified group, or slices",
		Run:   stopRun,
	}
)

func stopRun(cmd *cobra.Command, args []string) {
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
		newLogger.Error(nil, "%#v", maskAny(err))
		os.Exit(1)
	}

	taskObject, err := newController.Stop(req)
	if err != nil {
		newLogger.Error(nil, "%#v", maskAny(err))
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
