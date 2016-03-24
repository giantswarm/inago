package cli

import (
	"fmt"

	"github.com/juju/errgo"

	"github.com/giantswarm/inago/controller"
)

var (
	maskAny = errgo.MaskFunc(errgo.Any)
)

var invalidArgumentsError = errgo.Newf("invalid arguments")

// IsInvalidArgumentsError checks whether the given command line
// arguments are valid
func IsInvalidArgumentsError(err error) bool {
	return errgo.Cause(err) == invalidArgumentsError
}

// FormatValidationError returns the CausingErrors formatted:
// Validation Error found:
//		* unit slice not found
//		* unit does not have group prefix
func FormatValidationError(err controller.ValidationError) string {
	msg := "Validation Error found:\n"
	for _, err := range err.CausingErrors {
		msg = msg + fmt.Sprintf("\t* %v\n", err)
	}
	return msg
}
