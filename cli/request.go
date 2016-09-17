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
	var group string
	abs, err := filepath.Abs(groupdir)
	if err != nil {
		return controller.Request{}, maskAny(err)
	}
	group = filepath.Base(abs)

	config := controller.DefaultRequestConfig()
	config.Group = group
	req := controller.NewRequest(config)
	return filledWithUnits(req, abs, group)
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

func filledWithUnits(req controller.Request, dir, group string) (controller.Request, error) {
	finfos, err := ioutil.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return controller.Request{}, maskAny(groupNotExistError)
		}
		return controller.Request{}, maskAny(err)
	}
	if len(finfos) == 0 {
		return controller.Request{}, noUnitFilesError
	}
	req.Units = make([]controller.Unit, 0, 1)
	for _, f := range finfos {
		if f.IsDir() {
			continue
		}
		if !strings.HasPrefix(f.Name(), group) {
			continue
		}
		raw, err := ioutil.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			return controller.Request{}, maskAny(err)
		}
		req.Units = append(req.Units, controller.Unit{Name: f.Name(), Content: string(raw)})
	}
	return req, nil
}
