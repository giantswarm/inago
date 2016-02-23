package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	destroyCmd = &cobra.Command{
		Use:   "destroy [group [group..]]",
		Short: "Destroys the specified group or slices",
		Long:  "Destroys a group or the specified slices",
		Run:   destroyRun,
	}
)

func destroyRun(cmd *cobra.Command, args []string) {
	req, err := createRequest(args)
	if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}

	err = newController.Destroy(req)
	if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}

	if req.SliceIDs == nil {
		fmt.Printf("Destroyed group '%s'\n", req.Group)
	} else if len(req.SliceIDs) == 0 {
		fmt.Printf("Destroyed all slices of group '%s'\n", req.Group)
	} else {
		fmt.Printf("Destroyed %d slices for group '%s': %v", len(req.SliceIDs), req.Group, req.SliceIDs)
	}
}
