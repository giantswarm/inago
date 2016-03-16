package cli

import (
	"fmt"
	"os"

	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/giantswarm/inago/controller"
)

var (
	statusCmd = &cobra.Command{
		Use:   "status <group>",
		Short: "Get group status",
		Long:  "Print the status of a group",
		Run:   statusRun,
	}
)

func statusRun(cmd *cobra.Command, args []string) {
	newLogger.Debug(newCtx, "cli: starting status")

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
	handleStatusCmdError(newCtx, req, err)

	statusList, err := newController.GetStatus(newCtx, req)
	handleStatusCmdError(newCtx, req, err)

	data, err := createStatus(req.Group, statusList)
	handleStatusCmdError(newCtx, req, err)
	fmt.Println(columnize.SimpleFormat(data))
}

func handleStatusCmdError(ctx context.Context, req controller.Request, err error) {
	if controller.IsUnitNotFound(err) || controller.IsUnitSliceNotFound(err) {
		if req.SliceIDs == nil {
			newLogger.Error(ctx, "Failed to find group '%s'.", req.Group)
		} else if len(req.SliceIDs) == 0 {
			newLogger.Error(ctx, "Failed to find all slices of group '%s'.", req.Group)
		} else {
			newLogger.Error(ctx, "Failed to find %d slices for group '%s': %v.", len(req.SliceIDs), req.Group, req.SliceIDs)
		}
		os.Exit(1)
	} else if err != nil {
		newLogger.Error(ctx, "%#v", maskAny(err))
		os.Exit(1)
	}
}
