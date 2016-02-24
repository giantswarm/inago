// Package formicactl implements a command line client for formica. Cobra CLI
// is used as framework.
package cli

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/giantswarm/formica/controller"
	"github.com/giantswarm/formica/file-system/real"
	"github.com/giantswarm/formica/file-system/spec"
	"github.com/giantswarm/formica/fleet"
	"github.com/giantswarm/formica/task"
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
	newTask       task.Task

	MainCmd = &cobra.Command{
		Use:   "formicactl",
		Short: "orchestrate groups of unit files on Fleet clusters",
		Long:  "orchestrate groups of unit files on Fleet clusters",
		Run:   mainRun,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// This callback is executed after flags are parsed and before any
			// command runs.

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

			newTaskConfig := task.DefaultTaskConfig()
			newTask = task.NewTask(newTaskConfig)
		},
	}
)

func init() {
	MainCmd.PersistentFlags().StringVar(&globalFlags.FleetEndpoint, "fleet-endpoint", "unix:///var/run/fleet.sock", "endpoint used to connect to fleet")
	MainCmd.PersistentFlags().BoolVar(&globalFlags.NoBlock, "no-block", false, "block on syncronous actions or not")
	MainCmd.PersistentFlags().BoolVarP(&globalFlags.Verbose, "verbose", "v", false, "verbose output or not")

	MainCmd.AddCommand(submitCmd)
	MainCmd.AddCommand(statusCmd)
	MainCmd.AddCommand(startCmd)
	MainCmd.AddCommand(stopCmd)
	MainCmd.AddCommand(destroyCmd)
}

func mainRun(cmd *cobra.Command, args []string) {
	cmd.Help()
}
