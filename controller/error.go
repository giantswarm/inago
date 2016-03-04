package controller

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

var unitNotFoundError = errgo.New("unit not found")

// IsUnitNotFound checks whether the given error indicates the problem of an
// unit being found or not. In case you want to lookup the state of a group and
// no unit of it can be found on any machine, an error that you can identify
// using this method is returned.
func IsUnitNotFound(err error) bool {
	return errgo.Cause(err) == unitNotFoundError
}

var unitSliceNotFoundError = errgo.New("unit slice not found")

// IsUnitSliceNotFound checks whether the given error indicates the problem of
// an unit slice being found or not. In case you want to lookup the state of a
// group and some unit of this group cannot be found on any machine, an error
// that you can identify using this method is returned.
func IsUnitSliceNotFound(err error) bool {
	return errgo.Cause(err) == unitSliceNotFoundError
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

var waitTimeoutReachedError = errgo.New("wait timeout reached")

// IsWaitTimeoutReached asserts waitTimeoutReachedError.
func IsWaitTimeoutReached(err error) bool {
	return errgo.Cause(err) == waitTimeoutReachedError
}

var invalidArgumentError = errgo.Newf("invalid argument")

// IsInvalidArgument checks whether the given error indicates a invalid argument
// to the operation that was performed.
func IsInvalidArgument(err error) bool {
	return errgo.Cause(err) == invalidArgumentError
}

var updateNotAllowedError = errgo.Newf("update not allowed")

// IsUpdateNotAllowed asserts updateNotAllowedError.
func IsUpdateNotAllowed(err error) bool {
	return errgo.Cause(err) == updateNotAllowedError
}
