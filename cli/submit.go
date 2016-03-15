package cli

import (
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/giantswarm/inago/controller"
)

var (
	submitCmd = &cobra.Command{
		Use:   "submit <group> [scale]",
		Short: "Submit a group",
		Long:  "Submit a group to the cluster, with an optional scale",
		Run:   submitRun,
	}
)

func submitRun(cmd *cobra.Command, args []string) {
	group := ""
	scale := 1
	switch len(args) {
	case 1:
		group = args[0]
	case 2:
		group = args[0]
		n, err := strconv.Atoi(args[1])
		if err != nil {
			newLogger.Error(nil, "%#v\n", maskAny(err))
			os.Exit(1)
		}
		scale = n
	default:
		cmd.Help()
		os.Exit(1)
	}

	newRequestConfig := controller.DefaultRequestConfig()
	newRequestConfig.Group = group
	newRequestConfig.SliceIDs = strings.Split(strings.Repeat("x", scale), "")
	req := controller.NewRequest(newRequestConfig)

	req, err := extendRequestWithContent(fs, req)
	if err != nil {
		newLogger.Error(nil, "%#v\n", maskAny(err))
		os.Exit(1)
	}
	req, err = newController.ExtendWithRandomSliceIDs(req)
	if err != nil {
		newLogger.Error(nil, "%#v", maskAny(err))
		os.Exit(1)
	}

	taskObject, err := newController.Submit(req)
	if err != nil {
		newLogger.Error(nil, "%#v", maskAny(err))
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
