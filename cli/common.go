package cli

import (
	"fmt"
	"os"

	"github.com/giantswarm/inago/controller"
	"github.com/giantswarm/inago/task"
)

type blockWithFeedbackCtx struct {
	Request    controller.Request
	Descriptor string
	NoBlock    bool
	TaskID     string
	Closer     chan struct{}
}

func maybeBlockWithFeedback(ctx blockWithFeedbackCtx) {
	if !ctx.NoBlock {
		taskObject, err := newController.WaitForTask(ctx.TaskID, ctx.Closer)
		if err != nil {
			fmt.Printf("%#v\n", maskAny(err))
			os.Exit(1)
		}

		if task.HasFailedStatus(taskObject) {
			if ctx.Request.SliceIDs == nil {
				fmt.Printf("Failed to %s group '%s'. (%s)\n", ctx.Descriptor, ctx.Request.Group, taskObject.Error)
			} else if len(ctx.Request.SliceIDs) == 0 {
				fmt.Printf("Failed to %s all slices of group '%s'. (%s)\n", ctx.Descriptor, ctx.Request.Group, taskObject.Error)
			} else {
				fmt.Printf("Failed to %s %d slices for group '%s': %v. (%s)\n", ctx.Descriptor, len(ctx.Request.SliceIDs), ctx.Request.Group, ctx.Request.SliceIDs, taskObject.Error)
			}
			os.Exit(1)
		}
	}

	if ctx.Request.SliceIDs == nil {
		fmt.Printf("Succeeded to %s group '%s'.\n", ctx.Descriptor, ctx.Request.Group)
	} else if len(ctx.Request.SliceIDs) == 0 {
		fmt.Printf("Succeeded to %s all slices of group '%s'.\n", ctx.Descriptor, ctx.Request.Group)
	} else {
		fmt.Printf("Succeeded to %s %d slices for group '%s': %v.\n", ctx.Descriptor, len(ctx.Request.SliceIDs), ctx.Request.Group, ctx.Request.SliceIDs)
	}
}
