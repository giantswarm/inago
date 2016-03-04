// Package common provides implementation of general interest for certain sub
// packages of inago.
package common

import (
	"regexp"
)

// ExtExp matches unit file extensions.
//
//   app@1.service  =>  .service
//   app@1.mount    =>  .mount
//   app.service    =>  .service
//   app.mount      =>  .mount
//
var ExtExp = regexp.MustCompile(`(?m)\.[a-z]*$`)

// SliceID takes a unit file name and returns its slice ID.
//
//   app@1.service  =>  @1
//   app@1.mount    =>  @1
//   app.service    =>
//   app.mount      =>
//
func SliceID(name string) (string, error) {
	suffix, err := sliceSuffix(name)
	if err != nil {
		return "", maskAny(err)
	}
	ID := ExtExp.ReplaceAllString(suffix, "")

	return ID, nil
}

var groupExp = regexp.MustCompile("@(.*)")

func sliceSuffix(name string) (string, error) {
	found := groupExp.FindAllString(name, -1)
	if len(found) == 0 {
		return "", nil
	} else if len(found) > 1 {
		return "", maskAny(invalidArgumentsError)
	}
	return found[0], nil
}
