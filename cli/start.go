package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/giantswarm/formica/task"
)

var (
	startCmd = &cobra.Command{
		Use:   "start [group [group..]]",
		Short: "Starts the specified group or slices",
		Long:  "Starts a group or the specified slices",
		Run:   startRun,
	}
)

func startRun(cmd *cobra.Command, args []string) {
	req, err := createRequest(args)
	if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}

	action := func() error {
		err = newController.Start(req)
		if err != nil {
			return maskAny(err)
		}

		return nil
	}

	taskObject, err := newTask.Create(action)
	if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}

	if !globalFlags.NoBlock {
		taskObject, err = newTask.WaitForFinalStatus(taskObject, nil)
		if err != nil {
			fmt.Printf("%#v\n", maskAny(err))
			os.Exit(1)
		}

		if task.HasFailedStatus(taskObject) {
			if err != nil {
				if req.SliceIDs == nil {
					fmt.Printf("Failed to start group '%s'\n", req.Group)
				} else if len(req.SliceIDs) == 0 {
					fmt.Printf("Failed to start all slices of group '%s'\n", req.Group)
				} else {
					fmt.Printf("Failed to start %d slices for group '%s': %v", len(req.SliceIDs), req.Group, req.SliceIDs)
				}
				os.Exit(1)
			}
		}
	}

	if req.SliceIDs == nil {
		fmt.Printf("Started group '%s'\n", req.Group)
	} else if len(req.SliceIDs) == 0 {
		fmt.Printf("Started all slices of group '%s'\n", req.Group)
	} else {
		fmt.Printf("Started %d slices for group '%s': %v", len(req.SliceIDs), req.Group, req.SliceIDs)
	}
}
