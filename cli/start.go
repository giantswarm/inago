package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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

	err = newController.Start(req)
	if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}

	if req.SliceIDs == nil {
		fmt.Printf("Started group '%s'\n", req.Group)
	} else if len(req.SliceIDs) == 0 {
		fmt.Printf("Started all slices of group '%s'\n", req.Group)
	} else {
		fmt.Printf("Started %d slices for group '%s': %v", len(req.SliceIDs), req.Group, req.SliceIDs)
	}
}
