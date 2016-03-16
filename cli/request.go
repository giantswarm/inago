package cli

import (
	"path/filepath"

	"github.com/juju/errgo"

	"github.com/giantswarm/inago/controller"
	"github.com/giantswarm/inago/file-system/spec"
)

// readUnitFiles reads the given dir and returns a map of filename => filecontent.
// If any read operation fails, the error is immediately returned.
func readUnitFiles(fs filesystemspec.FileSystem, dir string) (map[string]string, error) {
	fileInfos, err := fs.ReadDir(dir)
	if err != nil {
		return nil, maskAny(err)
	}

	unitFiles := map[string]string{}
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
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
func extendRequestWithContent(fs filesystemspec.FileSystem, req controller.Request) (controller.Request, error) {
	unitFiles, err := readUnitFiles(fs, req.Group)
	if err != nil {
		return controller.Request{}, maskAny(err)
	}
	for name, content := range unitFiles {
		req.Units = append(req.Units, controller.Unit{Name: name, Content: content})
	}

	if len(req.Units) == 0 {
		return controller.Request{}, errgo.Newf("No unit files found for group '%s'", req.Group)
	}

	return req, nil
}
