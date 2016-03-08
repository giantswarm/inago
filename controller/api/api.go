// Package api defines the controller API interface.
package api

import (
	"fmt"
	"regexp"

	"github.com/giantswarm/inago/fleet"
	"github.com/giantswarm/inago/task"
)

// Controller defines the interface a controller needs to implement to provide
// operations for groups of unit files against a fleet cluster.
type Controller interface {
	// Submit schedules a group on the configured fleet cluster. This is done by
	// setting the state of the units in the group to loaded.
	Submit(req Request) (*task.Task, error)

	// Start starts a group on the configured fleet cluster. This is done by
	// setting the state of the units in the group to launched.
	Start(req Request) (*task.Task, error)

	// Stop stops a group on the configured fleet cluster. This is done by
	// setting the state of the units in the group to loaded.
	Stop(req Request) (*task.Task, error)

	// Destroy delets a group on the configured fleet cluster. This is done by
	// setting the state of the units in the group to inactive.
	Destroy(req Request) (*task.Task, error)

	// GetStatus fetches the current status of a group. If the unit cannot be
	// found, an error that you can identify using IsUnitNotFound is returned.
	GetStatus(req Request) ([]fleet.UnitStatus, error)

	// WaitForStatus waits for a group to reach the given status.
	WaitForStatus(req Request, desired Status, closer <-chan struct{}) error

	// WaitForTask waits for the given task to reach a final status. Once the
	// given task has reached the final status, the final task representation is
	// returned.
	WaitForTask(taskID string, closer <-chan struct{}) (*task.Task, error)
}

// Request represents a controller request. This is used to process some action
// on the controller.
type Request struct {
	// Group represents the plain group name without any slice expression.
	Group string

	// SliceIDs contains the IDs to create. IDs can be "1", "first", "whatever",
	// "5", etc..
	SliceIDs []string

	// Units represents a list of unit files that is supposed to be extended
	// using the provided slice IDs.
	Units []Unit
}

var unitExp = regexp.MustCompile("@.service")

// ExtendSlices extends unit files with respect to the given slice IDs. Having
// slice IDs "1" and "2" and having unit files "foo@.service" and
// "bar@.service" results in the following extended unit files.
//
// 	 foo@1.service
// 	 bar@1.service
// 	 foo@2.service
// 	 bar@2.service
//
func (r Request) ExtendSlices() (Request, error) {
	if len(r.SliceIDs) == 0 {
		return r, nil
	}
	newRequest := Request{
		SliceIDs: r.SliceIDs,
		Units:    []Unit{},
	}

	for _, sliceID := range r.SliceIDs {
		for _, unit := range r.Units {
			newUnit := unit
			newUnit.Name = unitExp.ReplaceAllString(newUnit.Name, fmt.Sprintf("@%s.service", sliceID))
			newRequest.Units = append(newRequest.Units, newUnit)
		}
	}

	return newRequest, nil
}

// Unit represents a systemd unit file.
type Unit struct {
	// Name is something like "appd@.service". It needs to be extended using the
	// slice ID before submitting to fleet.
	Name string

	// Content represents normal systemd unit file content.
	Content string
}

// Status represents the current status of a unit.
type Status string
