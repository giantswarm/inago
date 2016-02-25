package task

import (
	"github.com/juju/errgo"
)

var (
	maskAny = errgo.MaskFunc(errgo.Any)
)

var taskObjectNotFoundError = errgo.New("task object not found")

// IsTaskObjectNotFound checks whether the given error indicates the problem of
// an task object not being found or not. In case you want to lookup a task
// object that cannot be found in the underlying storage, an error that you can
// identify using this method is returned.
func IsTaskObjectNotFound(err error) bool {
	return errgo.Cause(err) == taskObjectNotFoundError
}
