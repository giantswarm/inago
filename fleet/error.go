package fleet

import (
	"github.com/juju/errgo"
)

var (
	maskAny = errgo.MaskFunc(errgo.Any)
)

var ipNotFoundError = errgo.New("ip not found")

// IsIPNotFound checks whether the given error indicates the problem of an IP
// not being found or not. In case you want to lookup the IP of a unit that
// cannot be found on any machine, this error is returned.
func IsIpNotFound(err error) bool {
	return errgo.Cause(err) == ipNotFoundError
}
