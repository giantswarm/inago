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
// cannot be found on any machine, an error that you can identify using this
// method is returned.
func IsIPNotFound(err error) bool {
	return errgo.Cause(err) == ipNotFoundError
}

var unitNotFoundError = errgo.New("unit not found")

// IsUnitNotFound checks whether the given error indicates the problem of an unit
// not being found or not. In case you want to lookup the state of a unit that
// cannot be found on any machine, an error that you can identify using this
// method is returned.
func IsUnitNotFound(err error) bool {
	return errgo.Cause(err) == unitNotFoundError
}

var invalidUnitStatusError = errgo.New("invalid unit status")

// IsInvalidUnitStatus checks whether the given error indicates the problem of
// an unexpected unit status response. In case you want to lookup the state of
// a unit that cannot be found on any machine or exists multiple times (what
// should never happen), an error that you can identify using this method is
// returned.
func IsInvalidUnitStatus(err error) bool {
	return errgo.Cause(err) == invalidUnitStatusError
}
