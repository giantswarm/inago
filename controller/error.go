package controller

import (
	"fmt"

	"github.com/juju/errgo"
)

// ValidationError capsules validation errors into one error struct.
// It is returned when the validation fails. causingErrors contains all
// errors that occurenced, while validating the request.
type ValidationError struct {
	causingErrors []error
}

func (e ValidationError) Error() string {
	// TODO better error message
	msg := "group or unit invalid:\n"
	for _, err := range e.causingErrors {
		msg = msg + fmt.Sprintf("\t%v\n", err)
	}
	return msg
}

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

var updateFailedError = errgo.Newf("update failed")

// IsUpdateFailed asserts updateFailedError.
func IsUpdateFailed(err error) bool {
	return errgo.Cause(err) == updateFailedError
}

var updateNotAllowedError = errgo.Newf("update not allowed")

// IsUpdateNotAllowed asserts updateNotAllowedError.
func IsUpdateNotAllowed(err error) bool {
	return errgo.Cause(err) == updateNotAllowedError
}

var noUnitsInGroupError = errgo.New("no units in group")

// IsNoUnitsInGroup returns true if the given error cause is noUnitsInGroupError.
func IsNoUnitsInGroup(err error) bool {
	return errgo.Cause(err) == noUnitsInGroupError
}

var badUnitPrefixError = errgo.New("unit does not have group prefix")

// IsBadUnitPrefix returns true if the given error cause is badUnitPrefixError.
func IsBadUnitPrefix(err error) bool {
	return errgo.Cause(err) == badUnitPrefixError
}

var mixedSliceInstanceError = errgo.New("group mixing scalable and non-scalable units")

// IsMixedSliceInstance returns true if the given error cause is mixedSliceInstanceError.
func IsMixedSliceInstance(err error) bool {
	return errgo.Cause(err) == mixedSliceInstanceError
}

var atInGroupNameError = errgo.New("@ symbols in group name")

// IsAtInGroupNameError returns true if the given error cause is atInGroupNameError.
func IsAtInGroupNameError(err error) bool {
	return errgo.Cause(err) == atInGroupNameError
}

var multipleAtInUnitNameError = errgo.New("multiple @ symbols in unit name")

// IsMultipleAtInUnitName returns true if the given error cause is multipleAtInUnitNameError.
func IsMultipleAtInUnitName(err error) bool {
	return errgo.Cause(err) == multipleAtInUnitNameError
}

var unitsSameNameError = errgo.New("unit named with same name as another unit")

// IsUnitsSameName returns true if the given error cause is unitsSameNameError.
func IsUnitsSameName(err error) bool {
	return errgo.Cause(err) == unitsSameNameError
}

var groupsArePrefixError = errgo.New("group is prefix of another group")

// IsGroupsArePrefix returns true if the given error cause is groupsArePrefixError.
func IsGroupsArePrefix(err error) bool {
	return errgo.Cause(err) == groupsArePrefixError
}

var groupsSameNameError = errgo.New("group named with same name as another group")

// IsGroupsSameName returns true if the given error cause is groupsSameNameError.
func IsGroupsSameName(err error) bool {
	return errgo.Cause(err) == groupsSameNameError
}

var invalidSubmitRequestSlicesGivenError = errgo.New("invalid submit request: slice ids given")

// InvalidSubmitRequestSlicesGiven returns true if the given error cause is invalidSubmitRequestSlicesGivenError.
func InvalidSubmitRequestSlicesGiven(err error) bool {
	return errgo.Cause(err) == invalidSubmitRequestSlicesGivenError
}
