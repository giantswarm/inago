package cli

import (
	"os"
	"strconv"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/giantswarm/inago/controller"
	"github.com/juju/errgo"
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
	newLogger.Debug(newCtx, "cli: starting submit")

	group := ""
	scale := 1
	switch len(args) {
	case 1:
		group = args[0]
	case 2:
		group = args[0]
		n, err := strconv.Atoi(args[1])
		if err != nil {
			newLogger.Error(newCtx, "%#v\n", maskAny(err))
			os.Exit(1)
		}
		scale = n
	default:
		cmd.Help()
		os.Exit(1)
	}

	req, err := createSubmitRequest(fs, group, scale)
	if err != nil {
		newLogger.Error(newCtx, "%#v", maskAny(err))
		os.Exit(1)
	}

	taskObject, err := newController.Submit(newCtx, req)
	if err != nil {
		newLogger.Error(newCtx, "%#v", maskAny(err))
		os.Exit(1)
	}

	maybeBlockWithFeedback(newCtx, blockWithFeedbackCtx{
		Request:    req,
		Descriptor: "submit",
		NoBlock:    globalFlags.NoBlock,
		TaskID:     taskObject.ID,
		Closer:     nil,
	})
}

func createSubmitRequest(fs afero.Afero, group string, scale int) (controller.Request, error) {
	newRequestConfig := controller.DefaultRequestConfig()
	newRequestConfig.Group = group

	req := controller.NewRequest(newRequestConfig)
	req, err := extendRequestWithContent(fs, req)
	if err != nil {
		return controller.Request{}, err
	}

	if strings.Contains(req.Units[0].Name, "@") {
		req.DesiredSlices = scale
	} else {
		if scale != 1 {
			return controller.Request{}, errgo.Newf("invalid scale: must be 1 for unscalable groups")
		}
		req.DesiredSlices = 1
	}
	req.SliceIDs = nil
	return req, nil
}
