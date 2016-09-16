package cli

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/juju/errgo"

	"github.com/giantswarm/inago/controller"
)

var (
	groupNotExistError = errgo.New("group does not exist")
	noUnitFilesError   = errgo.New("no unit files")
)

// readUnitFiles reads the given dir and returns a map of filename => filecontent.
// If any read operation fails, the error is immediately returned.
func readUnitFiles(dir string) (map[string]string, error) {
	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, maskAny(groupNotExistError)
		}
		return nil, maskAny(err)
	}

	unitFiles := map[string]string{}
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			continue
		}
		if !strings.HasPrefix(fileInfo.Name(), dir) {
			continue
		}

		raw, err := ioutil.ReadFile(filepath.Join(dir, fileInfo.Name()))
		if err != nil {
			return nil, maskAny(err)
		}

		unitFiles[fileInfo.Name()] = string(raw)
	}

	return unitFiles, nil
}

// extendRequestWithContent reads all unitfiles for the given group and returns
// a new Request with the Units filled.
func extendRequestWithContent(req controller.Request) (controller.Request, error) {
	unitFiles, err := readUnitFiles(req.Group)
	if err != nil {
		return controller.Request{}, maskAny(err)
	}
	for name, content := range unitFiles {
		req.Units = append(req.Units, controller.Unit{Name: name, Content: content})
	}

	if len(req.Units) == 0 {
		return controller.Request{}, noUnitFilesError
	}

	return req, nil
}

// parseGroupCLIArgs parses the given group arguments into a group and the
// given sliceIDs.
// "mygroup@123", "mygroup@456" => "mygroup", ["123", "456"]
func parseGroupCLIArgs(args []string) (string, []string, error) {
	group := strings.Split(args[0], "@")[0]
	sliceIDs := []string{}

	for _, arg := range args {
		split := strings.Split(arg, "@")
		// validate that groups are not mixed
		if split[0] != group {
			return "", nil, maskAny(invalidArgumentsError)
		}
		// only append slice ID if one was provided
		if len(split) > 1 {
			sliceIDs = append(sliceIDs, split[1])
		}
	}

	return group, sliceIDs, nil
}
