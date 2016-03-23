package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	projectVersion string

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Long:  "Print inagoctl version",
		Run:   versionRun,
	}
)

func versionRun(cmd *cobra.Command, args []string) {
	fmt.Printf("inagoctl %s\n", projectVersion)
}
