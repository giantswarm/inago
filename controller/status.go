package controller

import (
	"strings"

	"github.com/giantswarm/formica/common"
	"github.com/giantswarm/formica/fleet"
)

// UnitStatusList represents a list of UnitStatus.
type UnitStatusList []fleet.UnitStatus

// Group returns a shortened version of usl where equal status
// are replaced by one UnitStatus where the Name is replaced with "*".
func (usl UnitStatusList) Group() ([]fleet.UnitStatus, error) {
	matchers := map[string]struct{}{}
	newList := []fleet.UnitStatus{}

	for _, us := range usl {
		// Group unit status
		grouped, suffix, err := groupUnitStatus(usl, us)
		if err != nil {
			return nil, maskAny(err)
		}

		// Prevent doubled aggregation.
		if _, ok := matchers[suffix]; ok {
			continue
		}
		matchers[suffix] = struct{}{}

		// Aggregate.
		if allStatesEqual(grouped) {
			newStatus := grouped[0]
			newStatus.Name = "*"
			newList = append(newList, newStatus)
		} else {
			newList = append(newList, grouped...)
		}
	}

	return newList, nil
}

func groupUnitStatus(usl []fleet.UnitStatus, groupMember fleet.UnitStatus) ([]fleet.UnitStatus, string, error) {
	ID, err := common.SliceID(groupMember.Name)
	if err != nil {
		return nil, "", maskAny(invalidUnitStatusError)
	}

	newList := []fleet.UnitStatus{}
	for _, us := range usl {
		exp := common.ExtExp.ReplaceAllString(us.Name, "")
		if !strings.HasSuffix(exp, ID) {
			continue
		}

		newList = append(newList, us)
	}

	return newList, ID, nil
}

func allStatesEqual(usl []fleet.UnitStatus) bool {
	for _, us1 := range usl {
		for _, us2 := range usl {
			if us1.Current != us2.Current {
				return false
			}
			if us1.Desired != us2.Desired {
				return false
			}
			for _, m1 := range us1.Machine {
				for _, m2 := range us2.Machine {
					if m1.SystemdActive != m2.SystemdActive {
						return false
					}
				}
			}
		}
	}

	return true
}
