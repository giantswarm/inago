// Package cli implements a command line client for Inago. Cobra CLI
// is used as framework.
package cli

import (
	"net/url"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/giantswarm/inago/controller"
	"github.com/giantswarm/inago/file-system/real"
	"github.com/giantswarm/inago/file-system/spec"
	"github.com/giantswarm/inago/fleet"
	"github.com/giantswarm/inago/logging"
	"github.com/giantswarm/inago/task"
)

var (
	globalFlags struct {
		FleetEndpoint string
		NoBlock       bool
		Verbose       bool
	}

	fs             filesystemspec.FileSystem
	newLogger      logging.Logger
	newFleet       fleet.Fleet
	newTaskService task.Service
	newController  controller.Controller

	newCtx context.Context

	// MainCmd contains the cobra.Command to execute inagoctl.
	MainCmd = &cobra.Command{
		Use:   "inagoctl",
		Short: "Inago orchestrates groups of units on Fleet clusters",
		Long:  "Inago orchestrates groups of units on Fleet clusters",
		Run:   mainRun,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// This callback is executed after flags are parsed and before any
			// command runs.
			fs = filesystemreal.NewFileSystem()

			loggingConfig := logging.DefaultConfig()
			if globalFlags.Verbose {
				loggingConfig.LogLevel = "DEBUG"
			}
			newLogger = logging.NewLogger(loggingConfig)

			URL, err := url.Parse(globalFlags.FleetEndpoint)
			if err != nil {
				panic(err)
			}

			newFleetConfig := fleet.DefaultConfig()
			newFleetConfig.Endpoint = *URL
			newFleetConfig.Logger = newLogger
			newFleet, err = fleet.NewFleet(newFleetConfig)
			if err != nil {
				panic(err)
			}

			newTaskServiceConfig := task.DefaultConfig()
			newTaskServiceConfig.Logger = newLogger
			newTaskService = task.NewTaskService(newTaskServiceConfig)

			newControllerConfig := controller.DefaultConfig()
			newControllerConfig.Logger = newLogger
			newControllerConfig.Fleet = newFleet
			newControllerConfig.TaskService = newTaskService

			newController = controller.NewController(newControllerConfig)

			newCtx = context.Background()
		},
	}
)

func init() {
	MainCmd.PersistentFlags().StringVar(&globalFlags.FleetEndpoint, "fleet-endpoint", "unix:///var/run/fleet.sock", "endpoint used to connect to fleet")
	MainCmd.PersistentFlags().BoolVar(&globalFlags.NoBlock, "no-block", false, "block on synchronous actions")
	MainCmd.PersistentFlags().BoolVarP(&globalFlags.Verbose, "verbose", "v", false, "verbose output")

	MainCmd.AddCommand(submitCmd)
	MainCmd.AddCommand(statusCmd)
	MainCmd.AddCommand(startCmd)
	MainCmd.AddCommand(stopCmd)
	MainCmd.AddCommand(destroyCmd)
	MainCmd.AddCommand(updateCmd)
	MainCmd.AddCommand(validateCmd)
	MainCmd.AddCommand(versionCmd)
}

func mainRun(cmd *cobra.Command, args []string) {
	cmd.Help()
}
