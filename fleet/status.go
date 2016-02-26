package fleet

import (
	"strings"
)

type Status string

var (
	StatusFailed   Status = "failed"
	StatusRunning  Status = "running"
	StatusStarting Status = "starting"
	StatusStopped  Status = "stopped"
	StatusStopping Status = "stopping"
)

type StatusContext struct {
	FleetCurrent  string
	FleetDesired  string
	SystemdActive string
	SystemdSub    string
	Aggregated    Status
}

var (
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
// known to formica based on the StatusIndex.
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
