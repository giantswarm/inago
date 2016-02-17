package cli

import (
	"io/ioutil"
	"path/filepath"
	"regexp"

	"github.com/giantswarm/formica/controller"
	"github.com/giantswarm/formica/fleet"
)

var groupExp = regexp.MustCompile("@(.*)")

func readUnitFiles(dirName string) (map[string]string, error) {
	dirName = groupExp.ReplaceAllString(dirName, "")

	fileInfos, err := ioutil.ReadDir(dirName)
	if err != nil {
		return nil, maskAny(err)
	}

	unitFiles := map[string]string{}
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			continue
		}

		raw, err := ioutil.ReadFile(filepath.Join(dirName, fileInfo.Name()))
		if err != nil {
			return nil, maskAny(err)
		}

		unitFiles[fileInfo.Name()] = string(raw)
	}

	return unitFiles, nil
}

var atExp = regexp.MustCompile("@")

func createSliceIDs(dirName string) ([]string, error) {
	list := groupExp.FindAllString(dirName, -1)

	sliceIDs := []string{}
	for _, item := range list {
		sliceIDs = append(sliceIDs, atExp.ReplaceAllString(item, ""))
	}

	return sliceIDs, nil
}

func createRequest(dirName string) (controller.Request, error) {
	req := controller.Request{
		SliceIDs: []string{},
		Units:    []controller.Unit{},
	}

	unitFiles, err := readUnitFiles(dirName)
	if err != nil {
		return controller.Request{}, maskAny(err)
	}
	for name, content := range unitFiles {
		req.Units = append(req.Units, controller.Unit{Name: name, Content: content})
	}

	sliceIDs, err := createSliceIDs(dirName)
	if err != nil {
		return controller.Request{}, maskAny(err)
	}
	req.SliceIDs = sliceIDs

	return req, nil
}

func printStatus(groupStatus []fleet.UnitStatus) error {
	return nil
}
