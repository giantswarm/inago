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

// newRequestWithUnits reads the group directory's path for unit files and returns a copy of
// request filled with units and the group name set. The group name is the name of the gorup
// directory. An unit file is read only when it is not a directory and its name is prefixed
// with the group's name.
func newRequestWithUnits(groupdir string) (controller.Request, error) {
	group, err := base(groupdir)
	if err != nil {
		return controller.Request{}, maskAny(err)
	}
	units, err := readUnits(filepath.Clean(groupdir), group)
	if err != nil {
		return controller.Request{}, maskAny(err)
	}
	config := controller.DefaultRequestConfig()
	config.Group = group
	req := controller.NewRequest(config)
	req.Units = units
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

// base returns the last element of an absolute representation of path.
func base(path string) (string, error) {
	abs, err := filepath.Abs(path)
	return filepath.Base(abs), err
}

// readUnits read unit files form a given directory ensuring that there is at least one unit file.
// Unit file names are required to be prefixed by the group name.
func readUnits(dir, group string) ([]controller.Unit, error) {
	finfos, err := ioutil.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, maskAny(groupNotExistError)
		}
		return nil, maskAny(err)
	}
	if len(finfos) == 0 {
		return nil, noUnitFilesError
	}
	units := make([]controller.Unit, 0, 1)
	for _, f := range finfos {
		if f.IsDir() {
			continue
		}
		if !strings.HasPrefix(f.Name(), group) {
			continue
		}
		raw, err := ioutil.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			return nil, maskAny(err)
		}
		units = append(units, controller.Unit{Name: f.Name(), Content: string(raw)})
	}
	return units, nil
}
