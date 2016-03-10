// Package cli implements a command line client for Inago. Cobra CLI
// is used as framework.
package cli

import (
	"net/url"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/giantswarm/inago/controller"
	"github.com/giantswarm/inago/file-system/real"
	"github.com/giantswarm/inago/file-system/spec"
	"github.com/giantswarm/inago/fleet"
	"github.com/giantswarm/inago/logging"
)

var (
	globalFlags struct {
		FleetEndpoint string
		NoBlock       bool
		Verbose       bool
	}

	newController controller.Controller
	newFileSystem filesystemspec.FileSystem
	newFleet      fleet.Fleet

	// MainCmd contains the cobra.Command to execute inagoctl.
	MainCmd = &cobra.Command{
		Use:   "inagoctl",
		Short: "orchestrate groups of unit files on Fleet clusters",
		Long:  "orchestrate groups of unit files on Fleet clusters",
		Run:   mainRun,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// This callback is executed after flags are parsed and before any
			// command runs.

			if globalFlags.Verbose {
				logging.SetLogLevel("DEBUG")
			}

			if isatty.IsTerminal(os.Stderr.Fd()) {
				logging.UseColor(true)
			}

			URL, err := url.Parse(globalFlags.FleetEndpoint)
			if err != nil {
				panic(err)
			}

			newFleetConfig := fleet.DefaultConfig()
			newFleetConfig.Endpoint = *URL
			newFleet, err = fleet.NewFleet(newFleetConfig)
			if err != nil {
				panic(err)
			}

			newControllerConfig := controller.DefaultConfig()
			newControllerConfig.Fleet = newFleet
			newController = controller.NewController(newControllerConfig)

			newFileSystem = filesystemreal.NewFileSystem()
		},
	}
)

func init() {
	MainCmd.PersistentFlags().StringVar(&globalFlags.FleetEndpoint, "fleet-endpoint", "unix:///var/run/fleet.sock", "endpoint used to connect to fleet")
	MainCmd.PersistentFlags().BoolVar(&globalFlags.NoBlock, "no-block", false, "block on synchronous actions or not")
	MainCmd.PersistentFlags().BoolVarP(&globalFlags.Verbose, "verbose", "v", false, "verbose output or not")

	MainCmd.AddCommand(submitCmd)
	MainCmd.AddCommand(statusCmd)
	MainCmd.AddCommand(startCmd)
	MainCmd.AddCommand(stopCmd)
	MainCmd.AddCommand(destroyCmd)
	MainCmd.AddCommand(validateCmd)
}

func mainRun(cmd *cobra.Command, args []string) {
	cmd.Help()
}
