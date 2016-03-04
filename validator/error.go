package validator

import (
	"github.com/juju/errgo"
)

var (
	noUnitsInGroupError     = errgo.New("no units in group")
	badUnitPrefixError      = errgo.New("unit does not have group prefix")
	mixedSliceInstanceError = errgo.New("group mixing scalable and non-scalable units")

	groupsArePrefixError = errgo.New("group is prefix of another group")
)
