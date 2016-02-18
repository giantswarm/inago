package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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

	status, err := newController.GetStatus(req)
	if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}

	err = printStatus(status)
	if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}
}
