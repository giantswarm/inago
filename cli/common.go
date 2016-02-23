package cli

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/giantswarm/formica/controller"
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

func createRequestWithContent(slices []string) (controller.Request, error) {
	err := validateArgs(slices)
	if err != nil {
		return controller.Request{}, maskAny(err)
	}

	req := controller.Request{
		Group:    dirnameFromSlices(slices),
		SliceIDs: []string{},
		Units:    []controller.Unit{},
	}

	unitFiles, err := readUnitFiles(slices)
	if err != nil {
		return controller.Request{}, maskAny(err)
	}
	for name, content := range unitFiles {
		req.Units = append(req.Units, controller.Unit{Name: name, Content: content})
	}

	sliceIDs, err := createSliceIDs(slices)
	if err != nil {
		return controller.Request{}, maskAny(err)
	}
	req.SliceIDs = sliceIDs

	return req, nil
}

func createRequest(slices []string) (controller.Request, error) {
	err := validateArgs(slices)
	if err != nil {
		return controller.Request{}, maskAny(err)
	}

	req := controller.Request{
		Group:    dirnameFromSlices(slices),
		SliceIDs: []string{},
	}

	sliceIDs, err := createSliceIDs(slices)
	if err != nil {
		return controller.Request{}, maskAny(err)
	}
	req.SliceIDs = sliceIDs

	return req, nil
}

var (
	statusHeader = "Group | Units | FDState | FCState | SAState | IP | Machine"
	statusBody   = "%s | %s | %s | %s | %s | %s | %s"
)

func createStatus(group string, usl controller.UnitStatusList) ([]string, error) {
	data := []string{
		statusHeader,
		"",
	}

	usl, err := usl.Group()
	if err != nil {
		return nil, maskAny(err)
	}

	for _, us := range usl {
		for _, ms := range us.Machine {
			row := fmt.Sprintf(
				statusBody,
				group+us.Slice,
				us.Name,
				us.Desired,
				us.Current,
				ms.SystemdActive,
				ms.IP,
				ms.ID,
			)
			data = append(data, row)
		}
	}

	return data, nil
}
