package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"

	"github.com/giantswarm/formica/controller"
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
	req, err := createRequest(args)
	if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}

	group := dirnameFromSlices(args)
	statusList, err := newController.GetStatus(req)
	if controller.IsUnitNotFound(err) {
		fmt.Printf("No unit found for group slice(s) '%s'.\n", args)
		os.Exit(1)
	} else if controller.IsUnitSliceNotFound(err) {
		fmt.Printf("%s%s.\n", strings.ToUpper(err.Error()[0:1]), err.Error()[1:])
		os.Exit(1)
	} else if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}

	data, err := createStatus(group, statusList)
	if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}
	fmt.Println(columnize.SimpleFormat(data))
}
