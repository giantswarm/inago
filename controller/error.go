package controller

import (
	"fmt"

	"github.com/juju/errgo"
)

var (
	maskAny = errgo.MaskFunc(errgo.Any)
)

func maskAnyf(err error, f string, v ...interface{}) error {
	f = fmt.Sprintf("%s: %s", err.Error(), f)
	newErr := errgo.WithCausef(nil, errgo.Cause(err), f, v...)

	if e, _ := newErr.(*errgo.Err); e != nil {
		e.SetLocation(1)
		return e
	}

	return err
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
