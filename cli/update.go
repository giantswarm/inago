package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/giantswarm/inago/controller"
)

var (
	updateFlags struct {
		MaxGrowth int
		MinAlive  int
		ReadySecs int
	}

	updateCmd = &cobra.Command{
		Use:   "update [group]",
		Short: "update a group",
		Long:  "update a group",
		Run:   updateRun,
	}
)

func init() {
	updateCmd.PersistentFlags().IntVar(&updateFlags.MaxGrowth, "max-growth", 1, "maximum number of group slices added at a time")
	updateCmd.PersistentFlags().IntVar(&updateFlags.MinAlive, "min-alive", 1, "minimum number of group slices staying alive at a time")
	updateCmd.PersistentFlags().IntVar(&updateFlags.ReadySecs, "ready-secs", 30, "number of seconds to sleep before updating the next group slice")
}

func updateRun(cmd *cobra.Command, args []string) {
	group := ""
	switch len(args) {
	case 1:
		group = args[0]
	default:
		cmd.Help()
		os.Exit(1)
	}

	newRequestConfig := controller.DefaultRequestConfig()
	newRequestConfig.Group = group
	req := controller.NewRequest(newRequestConfig)

	req, err := newController.ExtendWithContent(req)
	handleUpdateCmdError(err)
	req, err = newController.ExtendWithExistingSliceIDs(req)
	handleUpdateCmdError(err)

	opts := controller.UpdateOptions{
		MaxGrowth: updateFlags.MaxGrowth,
		MinAlive:  updateFlags.MinAlive,
		ReadySecs: updateFlags.ReadySecs,

		// TODO Verbosity flag for displaying feedback about the current update steps?
		// TODO Force flag for forcing the update even if the unit hashes do not differ?
	}

	taskObject, err := newController.Update(req, opts)
	handleUpdateCmdError(err)
	taskObject, err = newController.WaitForTask(taskObject.ID, nil)
	handleUpdateCmdError(err)

	req, err = newController.ExtendWithExistingSliceIDs(req)
	handleUpdateCmdError(err)

	maybeBlockWithFeedback(blockWithFeedbackCtx{
		Request:    req,
		Descriptor: "update",
		NoBlock:    false,
		TaskID:     taskObject.ID,
		Closer:     nil,
	})
}

func handleUpdateCmdError(err error) {
	if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}
}
