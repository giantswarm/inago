// Package validator provides functionality for validating groups and units.
package validator

import (
	"sort"
	"strings"

	"github.com/giantswarm/inago/controller"
)

// StringsUnique returns true if all strings in the list are unique,
// false otherwise.
func StringsUnique(s []string) bool {
	sort.Strings(s)

	for i := 0; i < len(s)-1; i++ {
		if s[i] == s[i+1] {
			return false
		}
	}

	return true
}

// StringsHasPrefix returns true if all of the strings have the given prefix,
// false otherwise.
func StringsHasPrefix(s []string, p string) bool {
	for _, x := range s {
		if !strings.HasPrefix(x, p) {
			return false
		}
	}

	return true
}

// StringsSharePrefix returns true if any of the strings are prefixes of another,
// false otherwise.
func StringsSharePrefix(s []string) bool {
	sort.Strings(s)

	for i := 0; i < len(s)-1; i++ {
		if strings.HasPrefix(s[i+1], s[i]) {
			return true
		}
	}

	return false
}

// StringsCountMoreThan returns true if any of the strings in s
// contain more than n occurences of c, false otherwise.
func StringsCountMoreThan(s []string, c string, n int) bool {
	for _, x := range s {
		if strings.Count(x, c) > n {
			return true
		}
	}

	return false
}

// StringsHaveOrNot returns true if all strings in s either have an occurence of c,
// or do not have any occurence of c.
// In another way, it returns false if only some strings in s have an occurence of c.
func StringsHaveOrNot(s []string, c string) bool {
	numStringsWithOccurence := 0

	for _, x := range s {
		if strings.Contains(x, c) {
			numStringsWithOccurence++
		}
	}

	return !(numStringsWithOccurence > 0 && numStringsWithOccurence < len(s))
}

// ValidateRequest takes a Request, and returns whether it is valid or not.
// If the request is not valid, the error provides more details.
func ValidateRequest(request controller.Request) (bool, error) {
	// Check there are units in the group.
	if len(request.Units) == 0 {
		return false, noUnitsInGroupError
	}

	// Check that there are not any @ symbols in the group name.
	if strings.Contains(request.Group, "@") {
		return false, atInGroupNameError
	}

	unitNames := []string{}

	for _, unit := range request.Units {
		unitNames = append(unitNames, unit.Name)
	}

	// Check that we're not mixing units with @ and units without @.
	if !StringsHaveOrNot(unitNames, "@.") {
		return false, mixedSliceInstanceError
	}

	// Check that all unit names are prefixed by the group name.
	if !StringsHasPrefix(unitNames, request.Group) {
		return false, badUnitPrefixError
	}

	// Check that @ only occurences at most once per unit name.
	if StringsCountMoreThan(unitNames, "@", 1) {
		return false, multipleAtInUnitNameError
	}

	// Check that all unit names are unique.
	if !StringsUnique(unitNames) {
		return false, unitsSameNameError
	}

	return true, nil
}

// ValidateMultipleRequest takes a list of Requests, and returns whether
// they are valid together or not.
// If the requests are not valid, the error returned provides more details.
func ValidateMultipleRequest(requests []controller.Request) (bool, error) {
	groupNames := []string{}

	for _, request := range requests {
		groupNames = append(groupNames, request.Group)
	}

	// Check that all group names are unique.
	if !StringsUnique(groupNames) {
		return false, groupsSameNameError
	}

	// Check that group names are not prefixes of each other.
	if StringsSharePrefix(groupNames) {
		return false, groupsArePrefixError
	}

	return true, nil
}
