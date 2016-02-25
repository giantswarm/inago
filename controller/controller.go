// Package controller implements a controller client providing basic operations against a
// controller endpoint through controller's HTTP API. Higher level scheduling and
// management should be built on top of that.
package controller

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/giantswarm/formica/fleet"
	"github.com/giantswarm/formica/task"
)

// Config provides all necessary and injectable configurations for a new
// controller.
type Config struct {
	Fleet       fleet.Fleet
	TaskService task.TaskService
}

// DefaultConfig provides a set of configurations with default values by best
// effort.
func DefaultConfig() Config {
	newFleetConfig := fleet.DefaultConfig()
	newFleet, err := fleet.NewFleet(newFleetConfig)
	if err != nil {
		panic(err)
	}

	newTaskServiceConfig := task.DefaultTaskServiceConfig()
	newTaskService := task.NewTaskService(newTaskServiceConfig)

	newConfig := Config{
		Fleet:       newFleet,
		TaskService: newTaskService,
	}

	return newConfig
}

// Controller defines the interface a controller needs to implement to provide
// operations for groups of unit files against a fleet cluster.
type Controller interface {
	// Submit schedules a group on the configured fleet cluster. This is done by
	// setting the state of the units in the group to loaded.
	Submit(req Request) (*task.TaskObject, error)

	// Start starts a group on the configured fleet cluster. This is done by
	// setting the state of the units in the group to launched.
	Start(req Request) (*task.TaskObject, error)

	// Stop stops a group on the configured fleet cluster. This is done by
	// setting the state of the units in the group to loaded.
	Stop(req Request) (*task.TaskObject, error)

	// Destroy delets a group on the configured fleet cluster. This is done by
	// setting the state of the units in the group to inactive.
	Destroy(req Request) (*task.TaskObject, error)

	// GetStatus fetches the current status of a group. If the unit cannot be
	// found, an error that you can identify using IsUnitNotFound is returned.
	GetStatus(req Request) ([]fleet.UnitStatus, error)

	// WaitForTask waits for the given task to reach a final status. Once the
	// given task has reached the final status, the final task representation is
	// returned.
	WaitForTask(taskObject *task.TaskObject, closer <-chan struct{}) (*task.TaskObject, error)
}

// NewController creates a new Controller that is configured with the given
// settings.
//
//   newConfig := controller.DefaultConfig()
//   newConfig.Fleet = myCustomFleetClient
//   newController := controller.NewController(newConfig)
//
func NewController(config Config) Controller {
	newController := controller{
		Config: config,
	}

	return &newController
}

type controller struct {
	Config
}

// Unit represents a systemd unit file.
type Unit struct {
	// Name is something like "appd@.service". It needs to be extended using the
	// slice ID before submitting to fleet.
	Name string

	// Content represents normal systemd unit file content.
	Content string
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

func (c controller) Submit(req Request) (*task.TaskObject, error) {
	action := func() error {
		extended, err := req.ExtendSlices()
		if err != nil {
			return maskAny(err)
		}

		for _, unit := range extended.Units {
			err := c.Fleet.Submit(unit.Name, unit.Content)
			if err != nil {
				return maskAny(err)
			}
		}

		// TODO retry operations

		return nil
	}

	taskObject, err := c.TaskService.Create(action)
	if err != nil {
		return nil, maskAny(err)
	}

	return taskObject, nil
}

func (c controller) Start(req Request) (*task.TaskObject, error) {
	action := func() error {
		unitStatusList, err := c.groupStatus(req)
		if err != nil {
			return maskAny(err)
		}

		for _, unitStatus := range unitStatusList {
			err := c.Fleet.Start(unitStatus.Name)
			if err != nil {
				return maskAny(err)
			}
		}

		// TODO retry operations

		return nil
	}

	taskObject, err := c.TaskService.Create(action)
	if err != nil {
		return nil, maskAny(err)
	}

	return taskObject, nil
}

func (c controller) Stop(req Request) (*task.TaskObject, error) {
	action := func() error {
		unitStatusList, err := c.groupStatus(req)
		if err != nil {
			return maskAny(err)
		}

		for _, unitStatus := range unitStatusList {
			err := c.Fleet.Stop(unitStatus.Name)
			if err != nil {
				return maskAny(err)
			}
		}

		// TODO retry operations

		return nil
	}

	taskObject, err := c.TaskService.Create(action)
	if err != nil {
		return nil, maskAny(err)
	}

	return taskObject, nil
}

func (c controller) Destroy(req Request) (*task.TaskObject, error) {
	action := func() error {
		unitStatusList, err := c.groupStatus(req)
		if err != nil {
			return maskAny(err)
		}

		for _, unitStatus := range unitStatusList {
			err := c.Fleet.Destroy(unitStatus.Name)
			if err != nil {
				return maskAny(err)
			}
		}

		// TODO retry operations

		return nil
	}

	taskObject, err := c.TaskService.Create(action)
	if err != nil {
		return nil, maskAny(err)
	}

	return taskObject, nil
}

func (c controller) GetStatus(req Request) ([]fleet.UnitStatus, error) {
	status, err := c.groupStatus(req)
	return status, maskAny(err)
}

func (c controller) WaitForTask(taskObject *task.TaskObject, closer <-chan struct{}) (*task.TaskObject, error) {
	taskObject, err := c.TaskService.WaitForFinalStatus(taskObject, closer)
	return taskObject, maskAny(err)
}

func (c controller) groupStatus(req Request) ([]fleet.UnitStatus, error) {
	unitStatusList, err := c.Fleet.GetStatusWithMatcher(matchesGroupSlices(req))
	if fleet.IsUnitNotFound(err) {
		return []fleet.UnitStatus{}, maskAny(unitNotFoundError)
	} else if err != nil {
		return []fleet.UnitStatus{}, maskAny(err)
	}

	err = validateUnitStatusWithRequest(unitStatusList, req)
	if err != nil {
		return []fleet.UnitStatus{}, maskAny(err)
	}

	// TODO retry operations

	return unitStatusList, nil
}

func validateUnitStatusWithRequest(unitStatusList []fleet.UnitStatus, req Request) error {
	for _, sliceID := range req.SliceIDs {
		if !containsUnitStatusSliceID(unitStatusList, sliceID) {
			return maskAnyf(unitSliceNotFoundError, "slice ID '%s'", sliceID)
		}
	}

	return nil
}

func containsUnitStatusSliceID(unitStatusList []fleet.UnitStatus, sliceID string) bool {
	sliceID = fmt.Sprintf("@%s.service", sliceID)

	for _, us := range unitStatusList {
		if strings.HasSuffix(us.Name, sliceID) {
			return true
		}
	}

	return false
}

// matchesGroupSlices returns a matcher compatible with fleet.GetStatusWithMatcher
// that matches for each unitfiles that belongs to the group specified by
// request.Group and request.SliceIDs
func matchesGroupSlices(request Request) func(string) bool {
	// If only the group name is of interest, return shorter version
	if request.SliceIDs == nil || len(request.SliceIDs) == 0 {
		return func(name string) bool {
			return strings.HasPrefix(name, request.Group)
		}
	}

	// Normal version that matches on group prefix and slice ID suffix.
	return func(unitname string) bool {
		if !strings.HasPrefix(unitname, request.Group) {
			return false
		}

		for _, sliceID := range request.SliceIDs {
			if strings.HasSuffix(unitname, "@"+sliceID+".service") {
				return true
			}
		}
		return false
	}
}
