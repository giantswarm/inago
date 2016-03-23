package cli

import (
	"bytes"
	"net"
	"os"
	"strings"
	"text/template"

	"golang.org/x/net/context"

	"github.com/giantswarm/inago/controller"
	"github.com/giantswarm/inago/fleet"
	"github.com/giantswarm/inago/task"
)

var (
	statusHeader = "Group | Units | FDState | FCState | SAState {{if .Verbose}}| Hash {{end}}| IP | Machine"
	statusBody   = "{{.Group}}{{if .UnitState.SliceID}}@{{.UnitState.SliceID}}{{end}} | {{.UnitState.Name}} | {{.UnitState.Desired}} | {{.UnitState.Current}} | " +
		"{{.MachineState.SystemdActive}}{{if .Verbose}} | {{.MachineState.UnitHash}}{{end}} | {{if .MachineState.IP}}{{.MachineState.IP}}{{else}}-{{end}} | {{.MachineState.ID}}"
)

func createStatus(group string, usl controller.UnitStatusList) ([]string, error) {
	if !globalFlags.Verbose {
		var err error
		usl, err = usl.Group()
		if err != nil {
			return nil, maskAny(err)
		}
	}

	out := bytes.NewBufferString("")

	header := template.Must(template.New("header").Parse(statusHeader))
	header.Execute(out, struct {
		Verbose bool
	}{
		globalFlags.Verbose,
	})
	out.WriteString("\n\n")
	tmpl := template.Must(template.New("row-format").Parse(statusBody))

	addRow := func(group string, us, ms interface{}) {
		tmpl.Execute(out, struct {
			Verbose      bool
			Group        string
			UnitState    interface{}
			MachineState interface{}
		}{
			globalFlags.Verbose,
			group,
			us,
			ms,
		})
		out.WriteString("\n")
	}

	for _, us := range usl {
		if len(us.Machine) == 0 {
			addRow(group, us,
				fleet.MachineStatus{
					ID:            "-",
					IP:            net.IP{},
					SystemdActive: "-",
					SystemdSub:    "-",
					UnitHash:      "-",
				})
		}
		for _, ms := range us.Machine {
			addRow(group, us, ms)
		}
	}

	return strings.Split(out.String(), "\n"), nil
}

type blockWithFeedbackCtx struct {
	Request    controller.Request
	Descriptor string
	NoBlock    bool
	TaskID     string
	Closer     chan struct{}
}

func maybeBlockWithFeedback(ctx context.Context, bctx blockWithFeedbackCtx) {
	sliceNoun := "slices"
	if len(bctx.Request.SliceIDs) == 1 {
		sliceNoun = "slice"
	}

	if !bctx.NoBlock {
		taskObject, err := newController.WaitForTask(ctx, bctx.TaskID, bctx.Closer)
		if err != nil {
			newLogger.Error(ctx, "%#v", maskAny(err))
			os.Exit(1)
		}

		if controller.IsUnitsAlreadyUpToDate(taskObject.Error) {
			newLogger.Info(ctx, "Not updating group '%s'. (%s)", bctx.Request.Group, taskObject.Error.Error())
			return
		}

		if task.HasFailedStatus(taskObject) {
			if bctx.Request.SliceIDs == nil {
				newLogger.Error(ctx, "Failed to %s group '%s'. (%s)", bctx.Descriptor, bctx.Request.Group, taskObject.Error.Error())
			} else if len(bctx.Request.SliceIDs) == 0 {
				newLogger.Error(ctx, "Failed to %s all slices of group '%s'. (%s)", bctx.Descriptor, bctx.Request.Group, taskObject.Error.Error())
			} else {
				newLogger.Error(
					ctx,
					"Failed to %s %d %v for group '%s'. %v. (%s)",
					bctx.Descriptor,
					len(bctx.Request.SliceIDs),
					sliceNoun,
					bctx.Request.Group,
					bctx.Request.SliceIDs,
					taskObject.Error,
				)
			}
			os.Exit(1)
		}
	}

	if bctx.Request.SliceIDs == nil {
		newLogger.Info(ctx, "Succeeded to %s group '%s'.", bctx.Descriptor, bctx.Request.Group)
	} else if len(bctx.Request.SliceIDs) == 0 {
		newLogger.Info(ctx, "Succeeded to %s all slices of group '%s'.", bctx.Descriptor, bctx.Request.Group)
	} else {
		newLogger.Info(
			ctx,
			"Succeeded to %s %d %v for group '%s'. %v.",
			bctx.Descriptor,
			len(bctx.Request.SliceIDs),
			sliceNoun,
			bctx.Request.Group,
			bctx.Request.SliceIDs,
		)
	}
}
