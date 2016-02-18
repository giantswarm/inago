package cli

import (
	"path/filepath"
	"regexp"

	"github.com/giantswarm/formica/controller"
	"github.com/giantswarm/formica/fleet"
)

var groupExp = regexp.MustCompile("@(.*)")

func readUnitFiles(slices []string) (map[string]string, error) {
	slice := slices[0]
	dirName := groupExp.ReplaceAllString(slice, "")

	fileInfos, err := newFileSystem.ReadDir(dirName)
	if err != nil {
		return nil, maskAny(err)
	}

	unitFiles := map[string]string{}
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			continue
		}

		raw, err := newFileSystem.ReadFile(filepath.Join(dirName, fileInfo.Name()))
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

func createRequest(slices []string) (controller.Request, error) {
	err := validateArgs(slices)
	if err != nil {
		return controller.Request{}, maskAny(err)
	}

	req := controller.Request{
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

func printStatus(groupStatus []fleet.UnitStatus) error {
	return nil
}
