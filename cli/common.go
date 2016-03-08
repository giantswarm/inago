package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/giantswarm/inago/controller"
	"github.com/giantswarm/inago/controller/api"
	"github.com/giantswarm/inago/task"
)

var groupExp = regexp.MustCompile("@(.*)")

func dirnameFromSlices(slices []string) string {
	slice := slices[0]
	dirname := groupExp.ReplaceAllString(slice, "")
	return dirname
}

func readUnitFiles(slices []string) (map[string]string, error) {
	dirname := dirnameFromSlices(slices)

	fileInfos, err := newFileSystem.ReadDir(dirname)
	if err != nil {
		return nil, maskAny(err)
	}

	unitFiles := map[string]string{}
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			continue
		}

		raw, err := newFileSystem.ReadFile(filepath.Join(dirname, fileInfo.Name()))
		if err != nil {
			return nil, maskAny(err)
		}

		unitFiles[fileInfo.Name()] = string(raw)
	}

	return unitFiles, nil
}

var atExp = regexp.MustCompile("@")

func validateArgs(slices []string) error {
	if len(slices) == 0 {
		return maskAny(invalidArgumentsError)
	}

	baseSlice := slices[0]
	baseSlice = groupExp.ReplaceAllString(baseSlice, "")

	for _, slice := range slices {
		slice = groupExp.ReplaceAllString(slice, "")
		if slice != baseSlice {
			return maskAny(invalidArgumentsError)
		}
	}

	return nil
}

func createSliceIDs(slices []string) ([]string, error) {
	sliceIDs := []string{}
	for _, slice := range slices {
		found := groupExp.FindAllString(slice, -1)
		if len(found) > 1 {
			return nil, maskAny(invalidArgumentsError)
		}
		if len(found) == 0 {
			// When there is no slice expression "@" given, we are dealing with a
			// normal dirname, so there is no slice ID required.
			continue
		}
		sliceIDs = append(sliceIDs, atExp.ReplaceAllString(found[0], ""))
	}

	return sliceIDs, nil
}

func createRequestWithContent(slices []string) (api.Request, error) {
	err := validateArgs(slices)
	if err != nil {
		return api.Request{}, maskAny(err)
	}

	req := api.Request{
		Group:    dirnameFromSlices(slices),
		SliceIDs: []string{},
		Units:    []api.Unit{},
	}

	unitFiles, err := readUnitFiles(slices)
	if err != nil {
		return api.Request{}, maskAny(err)
	}
	for name, content := range unitFiles {
		req.Units = append(req.Units, api.Unit{Name: name, Content: content})
	}

	sliceIDs, err := createSliceIDs(slices)
	if err != nil {
		return api.Request{}, maskAny(err)
	}
	req.SliceIDs = sliceIDs

	return req, nil
}

func createRequest(slices []string) (api.Request, error) {
	err := validateArgs(slices)
	if err != nil {
		return api.Request{}, maskAny(err)
	}

	req := api.Request{
		Group:    dirnameFromSlices(slices),
		SliceIDs: []string{},
	}

	sliceIDs, err := createSliceIDs(slices)
	if err != nil {
		return api.Request{}, maskAny(err)
	}
	req.SliceIDs = sliceIDs

	return req, nil
}

var (
	statusHeader = "Group | Units | FDState | FCState | SAState {{if .Verbose}}| Hash {{end}}| IP | Machine"
	statusBody   = "{{.Group}}{{.UnitState.Slice}} | {{.UnitState.Name}} | {{.UnitState.Desired}} | {{.UnitState.Current}} | " +
		"{{.MachineState.SystemdActive}}{{if .Verbose}} | {{.MachineState.UnitHash}}{{end}} | {{.MachineState.IP}} | {{.MachineState.ID}}"
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
	for _, us := range usl {
		for _, ms := range us.Machine {

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
	}

	return strings.Split(out.String(), "\n"), nil
}

type blockWithFeedbackCtx struct {
	Request    api.Request
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
