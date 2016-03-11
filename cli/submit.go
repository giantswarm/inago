package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/giantswarm/inago/controller"
)

var (
	submitCmd = &cobra.Command{
		Use:   "submit <group> [scale]",
		Short: "submit a group",
		Long:  "submit a group",
		Run:   submitRun,
	}
)

func submitRun(cmd *cobra.Command, args []string) {
	group := ""
	scale := 1
	switch len(args) {
	case 1:
		group = args[0]
	case 2:
		group = args[0]
		n, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Printf("%#v\n", maskAny(err))
			os.Exit(1)
		}
		scale = n
	default:
		cmd.Help()
		os.Exit(1)
	}

	newRequestConfig := controller.DefaultRequestConfig()
	newRequestConfig.Group = group
	newRequestConfig.SliceIDs = strings.Split(strings.Repeat("x", scale), "")
	req := controller.NewRequest(newRequestConfig)

	req, err := newController.ExtendWithContent(req)
	if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}
	req, err = newController.ExtendWithRandomSliceIDs(req)
	if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}

	taskObject, err := newController.Submit(req)
	if err != nil {
		fmt.Printf("%#v\n", maskAny(err))
		os.Exit(1)
	}

	maybeBlockWithFeedback(blockWithFeedbackCtx{
		Request:    req,
		Descriptor: "submit",
		NoBlock:    globalFlags.NoBlock,
		TaskID:     taskObject.ID,
		Closer:     nil,
	})
}
