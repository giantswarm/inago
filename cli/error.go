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
