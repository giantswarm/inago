// Package cli implements a command line client for Inago. Cobra CLI
// is used as framework.
package cli

import (
	"net/url"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/giantswarm/inago/controller"
	"github.com/giantswarm/inago/fleet"
	"github.com/giantswarm/inago/logging"
	"github.com/giantswarm/inago/task"
)

var (
	globalFlags struct {
		FleetEndpoint string
		NoBlock       bool
		Verbose       bool

		Tunnel                   string
		SSHUsername              string
		SSHTimeout               time.Duration
		SSHStrictHostKeyChecking bool
		SSHKnownHostsFile        string
	}

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
			if globalFlags.Tunnel != "" {
				newSSHTunnelConfig := fleet.DefaultSSHTunnelConfig()
				newSSHTunnelConfig.Endpoint = *URL
				newSSHTunnelConfig.KnownHostsFile = globalFlags.SSHKnownHostsFile
				newSSHTunnelConfig.Logger = newLogger
				newSSHTunnelConfig.StrictHostKeyChecking = globalFlags.SSHStrictHostKeyChecking
				newSSHTunnelConfig.Timeout = globalFlags.SSHTimeout
				newSSHTunnelConfig.Tunnel = globalFlags.Tunnel
				newSSHTunnelConfig.Username = globalFlags.SSHUsername
				newSSHTunnel, err := fleet.NewSSHTunnel(newSSHTunnelConfig)
				if err != nil {
					panic(err)
				}
				newFleetConfig.SSHTunnel = newSSHTunnel
			}
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

	MainCmd.PersistentFlags().StringVar(&globalFlags.Tunnel, "tunnel", "", "use a tunnel to communicate with fleet")
	MainCmd.PersistentFlags().StringVar(&globalFlags.SSHUsername, "ssh-username", "core", "username to use when connecting to CoreOS machine")
	MainCmd.PersistentFlags().DurationVar(&globalFlags.SSHTimeout, "ssh-timeout", time.Duration(10*time.Second), "timeout in seconds when establishing the connection via SSH")
	MainCmd.PersistentFlags().BoolVar(&globalFlags.SSHStrictHostKeyChecking, "ssh-strict-host-key-checking", true, "verify host keys presented by remote machines before initiating SSH connections")
	MainCmd.PersistentFlags().StringVar(&globalFlags.SSHKnownHostsFile, "ssh-known-hosts-file", "~/.fleetctl/known_hosts", "file used to store remote machine fingerprints")

	MainCmd.AddCommand(submitCmd)
	MainCmd.AddCommand(statusCmd)
	MainCmd.AddCommand(startCmd)
	MainCmd.AddCommand(stopCmd)
	MainCmd.AddCommand(destroyCmd)
	MainCmd.AddCommand(upCmd)
	MainCmd.AddCommand(updateCmd)
	MainCmd.AddCommand(validateCmd)
	MainCmd.AddCommand(versionCmd)
}

func mainRun(cmd *cobra.Command, args []string) {
	cmd.Help()
}
