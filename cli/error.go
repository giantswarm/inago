package cli

import (
	"github.com/juju/errgo"
)

var (
	maskAny = errgo.MaskFunc(errgo.Any)
)

var invalidArgumentsError = errgo.Newf("invalid arguments")
