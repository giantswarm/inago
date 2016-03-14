package cli

import (
	"fmt"
	"os"

	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"

	"github.com/giantswarm/inago/controller"
)

var (
	statusCmd = &cobra.Command{
		Use:   "status [group-slice] ...",
		Short: "status of a group",
		Long:  "status of a group",
		Run:   statusRun,
	}
)

func statusRun(cmd *cobra.Command, args []string) {
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
	handleStatusCmdError(req, err)

	statusList, err := newController.GetStatus(req)
	handleStatusCmdError(req, err)

	data, err := createStatus(req.Group, statusList)
	handleStatusCmdError(req, err)
	fmt.Println(columnize.SimpleFormat(data))
}

func handleStatusCmdError(req controller.Request, err error) {
	if controller.IsUnitNotFound(err) || controller.IsUnitSliceNotFound(err) {
		if req.SliceIDs == nil {
			newLogger.Error(nil, "Failed to find group '%s'.", req.Group)
		} else if len(req.SliceIDs) == 0 {
			newLogger.Error(nil, "Failed to find all slices of group '%s'.", req.Group)
		} else {
			newLogger.Error(nil, "Failed to find %d slices for group '%s': %v.", len(req.SliceIDs), req.Group, req.SliceIDs)
		}
		os.Exit(1)
	} else if err != nil {
		newLogger.Error(nil, "%#v", maskAny(err))
		os.Exit(1)
	}
}
