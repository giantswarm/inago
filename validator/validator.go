// Package validator provides functionality for validating groups and units.
package validator

import (
	"sort"
	"strings"

	"github.com/giantswarm/inago/controller"
)

// ValidateRequest takes a Request, and returns whether it is valid or not.
// If the request is not valid, the error provides more details.
func ValidateRequest(request controller.Request) (bool, error) {
	// Check there are units in the group.
	if len(request.Units) == 0 {
		return false, noUnitsInGroupError
	}

	// Check that each unit name is prefixed with the group name.
	for _, unit := range request.Units {
		if !strings.HasPrefix(unit.Name, request.Group) {
			return false, badUnitPrefixError
		}
	}

	// Check that all units either have @ or they don't.
	numUnitsWithAtSymbol := 0
	for _, unit := range request.Units {
		if strings.Contains(unit.Name, "@.") {
			numUnitsWithAtSymbol++
		}
	}
	if numUnitsWithAtSymbol > 0 && numUnitsWithAtSymbol < len(request.Units) {
		return false, mixedSliceInstanceError
	}

	// Test there are not any @ symbols in the group name.
	if strings.Contains(request.Group, "@") {
		return false, atInGroupNameError
	}

	// Test there are not multiple @ symbols in any unit name.
	for _, unit := range request.Units {
		if strings.Count(unit.Name, "@") > 1 {
			return false, multipleAtInUnitNameError
		}
	}

	unitNames := []string{}

	for _, unit := range request.Units {
		unitNames = append(unitNames, unit.Name)
	}

	sort.Strings(unitNames)

	// Test that all unit names are unique
	for i := 0; i < len(unitNames)-1; i++ {
		if unitNames[i] == unitNames[i+1] {
			return false, unitsSameNameError
		}
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

	sort.Strings(groupNames)

	for i := 0; i < len(groupNames)-1; i++ {
		if groupNames[i] == groupNames[i+1] {
			return false, groupsSameNameError
		}

		if strings.HasPrefix(groupNames[i+1], groupNames[i]) {
			return false, groupsArePrefixError
		}
	}

	return true, nil
}
