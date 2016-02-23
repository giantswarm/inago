package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	stopCmd = &cobra.Command{
		Use:   "stop [group [group..]]",
		Short: "Stops the specified group or slices",
		Long:  "Stops a group or the specified slices",
		Run:   stopRun,
	}
)

func stopRun(cmd *cobra.Command, args []string) {
	req, err := createRequest(args)
	if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}

	err = newController.Stop(req)
	if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}

	if req.SliceIDs == nil {
		fmt.Printf("Stopped group '%s'\n", req.Group)
	} else if len(req.SliceIDs) == 0 {
		fmt.Printf("Stopped all slices of group '%s'\n", req.Group)
	} else {
		fmt.Printf("Stopped %d slices for group '%s': %v", len(req.SliceIDs), req.Group, req.SliceIDs)
	}
}
