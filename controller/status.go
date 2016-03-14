package controller

import (
	"strings"

	"github.com/giantswarm/inago/common"
	"github.com/giantswarm/inago/fleet"
)

// UnitStatusList represents a list of UnitStatus.
type UnitStatusList []fleet.UnitStatus

// Group returns a shortened version of usl where equal status
// are replaced by one UnitStatus where the Name is replaced with "*".
func (usl UnitStatusList) Group() (UnitStatusList, error) {
	matchers := map[string]struct{}{}
	newList := []fleet.UnitStatus{}

	hashesEqual, err := allHashesEqual(usl)
	if err != nil {
		return nil, maskAny(err)
	}

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

		statesEqual := allStatesEqual(grouped)

		// Aggregate.
		if hashesEqual && statesEqual {
			newStatus := grouped[0]
			newStatus.Name = "*"
			newList = append(newList, newStatus)
		} else {
			newList = append(newList, grouped...)
		}
	}

	return newList, nil
}

func (usl UnitStatusList) unitStatusesBySliceID(sliceID string) UnitStatusList {
	var newList []fleet.UnitStatus

	for _, us := range usl {
		if us.SliceID == sliceID {
			newList = append(newList, us)
		}
	}

	return newList
}

// groupUnitStatus returns a subset of usl where the sliceID matches the sliceID
// of groupMember, ignoring the unit names and extension.
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

// allHashesEqual is supposed to receive a list of unit statuses that is not
// grouped. This is necessary to compare unit hashes across groups.
func allHashesEqual(usl []fleet.UnitStatus) (bool, error) {
	uhis, err := groupUnitHashInfos(usl)
	if err != nil {
		return false, maskAny(err)
	}

	for _, uhi1 := range uhis {
		for _, uhi2 := range uhis {
			if uhi1.Base != uhi2.Base {
				continue
			}
			if uhi1.Hash != uhi2.Hash {
				return false, nil
			}
		}
	}

	return true, nil
}

type unitHashInfo struct {
	Base    string
	SliceID string
	Hash    string
}

func groupUnitHashInfos(usl []fleet.UnitStatus) ([]unitHashInfo, error) {
	var uhis []unitHashInfo

	for _, us1 := range usl {
		for _, us2 := range usl {
			if common.UnitBase(us1.Name) != common.UnitBase(us2.Name) {
				continue
			}
			for _, m1 := range us1.Machine {
				sliceID, err := common.SliceID(us1.Name)
				if err != nil {
					return nil, maskAny(err)
				}
				uhi := unitHashInfo{
					Base:    common.UnitBase(us1.Name),
					SliceID: sliceID,
					Hash:    m1.UnitHash,
				}
				uhis = append(uhis, uhi)
			}
		}
	}

	return uhis, nil
}

// allStatesEqual returns true if all elements in usl match for the following
// fields: Current, Desired, Machine.SystemdActive. Note this does not compare
// hashes sinces this method is supposed to receive only grouped unit statuses.
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

func unitHasStatus(us fleet.UnitStatus, status Status) (bool, error) {
	for _, ms := range us.Machine {
		aggregated, err := AggregateStatus(us.Current, us.Desired, ms.SystemdActive, ms.SystemdSub)
		if err != nil {
			return false, maskAny(err)
		}

		if aggregated != status {
			return false, nil
		}
	}

	return true, nil
}

// Status represents the current status of a unit.
type Status string

var (
	// StatusFailed represents a unit being failed.
	StatusFailed Status = "failed"

	// StatusNotFound represents a unit not being found.
	StatusNotFound Status = "not-found"

	// StatusRunning represents a unit running.
	StatusRunning Status = "running"

	// StatusStarting represents a unit starting.
	StatusStarting Status = "starting"

	// StatusStopped represents a unit that has stopped.
	StatusStopped Status = "stopped"

	// StatusStopping represents a unit stopping.
	StatusStopping Status = "stopping"
)

// StatusContext represents a units status from fleet and systemd.
type StatusContext struct {
	FleetCurrent  string
	FleetDesired  string
	SystemdActive string
	SystemdSub    string
	Aggregated    Status
}

var (
	// StatusIndex represents the aggregated states of a unit.
	StatusIndex = []StatusContext{
		{
			FleetCurrent:  "inactive",
			FleetDesired:  "*",
			SystemdActive: "*",
			SystemdSub:    "*",
			Aggregated:    StatusStopped,
		},
		{
			FleetCurrent:  "loaded|launched",
			FleetDesired:  "*",
			SystemdActive: "inactive",
			SystemdSub:    "*",
			Aggregated:    StatusStopped,
		},
		{
			FleetCurrent:  "loaded|launched",
			FleetDesired:  "*",
			SystemdActive: "failed",
			SystemdSub:    "*",
			Aggregated:    StatusFailed,
		},
		{
			FleetCurrent:  "loaded|launched",
			FleetDesired:  "*",
			SystemdActive: "activating",
			SystemdSub:    "*",
			Aggregated:    StatusStarting,
		},
		{
			FleetCurrent:  "loaded|launched",
			FleetDesired:  "*",
			SystemdActive: "deactivating",
			SystemdSub:    "*",
			Aggregated:    StatusStopping,
		},
		{
			FleetCurrent:  "loaded|launched",
			FleetDesired:  "*",
			SystemdActive: "active|reloading",
			SystemdSub:    "stop-sigterm|stop-post|stop",
			Aggregated:    StatusStopping,
		},
		{
			FleetCurrent:  "loaded|launched",
			FleetDesired:  "*",
			SystemdActive: "active|reloading",
			SystemdSub:    "auto-restart|launched|start-pre|start-pre|start-post|start|dead",
			Aggregated:    StatusStarting,
		},
		{
			FleetCurrent:  "loaded|launched",
			FleetDesired:  "*",
			SystemdActive: "active|reloading",
			SystemdSub:    "exited|running",
			Aggregated:    StatusRunning,
		},
	}
)

// AggregateStatus aggregates the given fleet and systemd states to a Status
// known to Inago based on the StatusIndex.
//
//   fc: fleet current state
//   fd: fleet desired state
//   sa: systemd active state
//   ss: systemd sub state
//
func AggregateStatus(fc, fd, sa, ss string) (Status, error) {
	for _, statusContext := range StatusIndex {
		if !matchState(statusContext.FleetCurrent, fc) {
			continue
		}
		if !matchState(statusContext.FleetDesired, fd) {
			continue
		}
		if !matchState(statusContext.SystemdActive, sa) {
			continue
		}
		if !matchState(statusContext.SystemdSub, ss) {
			continue
		}

		// All requirements matched, so return the aggregated status.
		return statusContext.Aggregated, nil
	}

	return "", maskAnyf(invalidUnitStatusError, "fc: %s, fd: %s, sa: %s, ss: %s", fc, fd, sa, ss)
}

func matchState(indexed, remote string) bool {
	if indexed == "*" {
		// When the indexed state is "*", we accept all states.
		return true
	}

	for _, splitted := range strings.Split(indexed, "|") {
		if splitted == remote {
			// When the indexed state is equal to the remote state, we accept it.
			return true
		}
	}

	// The given remote state does not match the criteria of the indexed state.
	return false
}
