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

	if pe, ok := cause.(*os.PathError); ok {
		return pe.Err == noSuchFileOrDirectoryError
	}

	return cause == noSuchFileOrDirectoryError
}

var invalidImplementationError = errgo.New("invalid implementation")

// IsInvalidImplementation checks for the given error to be
// invalidImplementationError.
func IsInvalidImplementation(err error) bool {
	return errgo.Cause(err) == invalidImplementationError
}

var notADirectoryError = errgo.New("not a directory")

// IsNotADirectory checks for the given error to be notADirectoryError.
func IsNotADirectory(err error) bool {
	cause := errgo.Cause(err)

	if sce, ok := cause.(*os.SyscallError); ok {
		return sce.Err == notADirectoryError
	}

	return cause == notADirectoryError
}
