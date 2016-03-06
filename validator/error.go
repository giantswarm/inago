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

	groupsArePrefixError = errgo.New("group is prefix of another group")
	groupsSameNameError  = errgo.New("group named with same name as another group")
)
