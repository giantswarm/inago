package fleet

import (
	"fmt"

	"github.com/juju/errgo"
)

var (
	maskAny = errgo.MaskFunc(errgo.Any)
)

// maskAnyf returns a new github.com/juju/errgo error wrapping the given one.
// The message will contain the message of f and v (see fmt.Printf), prefixed
// with the message of err.
//
// Examples:
//   maskAnyf(unitNotFoundError, "%s", unit.ID) => "unit not found: 12345abcdef.service"
func maskAnyf(err error, f string, v ...interface{}) error {
	if err == nil {
		return nil
	}

	f = fmt.Sprintf("%s: %s", err.Error(), f)
	newErr := errgo.WithCausef(nil, errgo.Cause(err), f, v...)
	newErr.(*errgo.Err).SetLocation(1)

	return newErr
}

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
