package validator

import (
	"github.com/juju/errgo"
)

var (
	noUnitsInGroupError        = errgo.New("no units in group")
	badUnitPrefixError         = errgo.New("unit does not have group prefix")
	mixedSliceInstanceError    = errgo.New("group mixing scalable and non-scalable units")
	multipleAtInGroupNameError = errgo.New("multiple @ symbols in group name")
	multipleAtInUnitNameError  = errgo.New("multiple @ symbols in unit name")
	unitsSameNameError         = errgo.New("unit named with same name as another unit")

	groupsArePrefixError = errgo.New("group is prefix of another group")
	groupsSameNameError  = errgo.New("group named with same name as another group")
)

// IsNoUnitsInGroup returns true if the given error cause is noUnitsInGroupError.
func IsNoUnitsInGroup(err error) bool {
	return errgo.Cause(err) == noUnitsInGroupError
}

// IsBadUnitPrefix returns true if the given error cause is badUnitPrefixError.
func IsBadUnitPrefix(err error) bool {
	return errgo.Cause(err) == badUnitPrefixError
}

// IsMixedSliceInstance returns true if the given error cause is mixedSliceInstanceError.
func IsMixedSliceInstance(err error) bool {
	return errgo.Cause(err) == mixedSliceInstanceError
}

// IsMultipleAtInGroupName returns true if the given error cause is multipleAtInGroupNameError.
func IsMultipleAtInGroupName(err error) bool {
	return errgo.Cause(err) == multipleAtInGroupNameError
}

// IsMultipleAtInUnitName returns true if the given error cause is multipleAtInUnitNameError.
func IsMultipleAtInUnitName(err error) bool {
	return errgo.Cause(err) == multipleAtInUnitNameError
}

// IsUnitsSameName returns true if the given error cause is unitsSameNameError.
func IsUnitsSameName(err error) bool {
	return errgo.Cause(err) == unitsSameNameError
}

// IsGroupsArePrefix returns true if the given error cause is groupsArePrefixError.
func IsGroupsArePrefix(err error) bool {
	return errgo.Cause(err) == groupsArePrefixError
}

// IsGroupsSameName returns true if the given error cause is groupsSameNameError.
func IsGroupsSameName(err error) bool {
	return errgo.Cause(err) == groupsSameNameError
}
