package cli

import (
	"fmt"
	"os"

	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"

	"github.com/giantswarm/inago/controller"
)

var (
	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all groups within a fleet cluster",
		Long:  "List all groups within a fleet cluster",
		Run:   listRun,
	}
)

func listRun(cmd *cobra.Command, args []string) {
	statusList, err := newController.List()
	if controller.IsUnitNotFound(err) || controller.IsUnitSliceNotFound(err) {
		fmt.Printf("No groups to show.\n")
		os.Exit(1)
	} else if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}

	data, err := createList(statusList)
	if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}
	fmt.Println(columnize.SimpleFormat(data))
}
