package cli

import (
	"path/filepath"
	"strings"

	"github.com/juju/errgo"
	"github.com/spf13/afero"

	"github.com/giantswarm/inago/controller"
)

var noUnitFilesError = errgo.New("no unit files")

// readUnitFiles reads the given dir and returns a map of filename => filecontent.
// If any read operation fails, the error is immediately returned.
func readUnitFiles(fs afero.Afero, dir string) (map[string]string, error) {
	fileInfos, err := fs.ReadDir(dir)
	if err != nil {
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

		raw, err := fs.ReadFile(filepath.Join(dir, fileInfo.Name()))
		if err != nil {
			return nil, maskAny(err)
		}

		unitFiles[fileInfo.Name()] = string(raw)
	}

	return unitFiles, nil
}

// extendRequestWithContent reads all unitfiles for the given group and returns
// a new Request with the Units filled.
func extendRequestWithContent(fs afero.Afero, req controller.Request) (controller.Request, error) {
	unitFiles, err := readUnitFiles(fs, req.Group)
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
