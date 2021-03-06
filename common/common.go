// Package common provides implementation of general interest for certain sub
// packages of Inago.
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
func SliceID(name string) (string, error) {
	suffix, err := sliceSuffix(name)
	if err != nil {
		return "", maskAny(err)
	}
	ID := ExtExp.ReplaceAllString(suffix, "")

	if ID == "" {
		return ID, nil
	}

	// Finally strip the @
	return ID[1:], nil
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

// UnitBase returns the base of the unit name. The base is considered
// everything up to the @ character, if any, or the given name without
// extension.
func UnitBase(name string) string {
	name = groupExp.ReplaceAllString(name, "")
	return ExtExp.ReplaceAllString(name, "")
}
