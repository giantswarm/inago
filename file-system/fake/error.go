package filesystemfake

import (
	"os"

	"github.com/juju/errgo"
)

var (
	maskAny = errgo.MaskFunc(errgo.Any)
)

var noSuchFileOrDirectoryError = errgo.New("no such file or directory")

// IsNoSuchFileOrDirectory checks for the given error to be
// noSuchFileOrDirectoryError. This error is returned in case there cannot any
// file be found as requested.
func IsNoSuchFileOrDirectory(err error) bool {
	cause := errgo.Cause(err)

	if cause == nil {
		return false
	}

	if pe, ok := cause.(*os.PathError); ok {
		return pe.Err == noSuchFileOrDirectoryError
	}

	return errgo.Cause(err) == noSuchFileOrDirectoryError
}
