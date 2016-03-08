// Package controller implements a controller client providing basic operations against a
// controller endpoint through controller's HTTP API. Higher level scheduling and
// management should be built on top of that.
package controller

import (
	"fmt"
	"strings"
	"time"

	"github.com/giantswarm/inago/controller/api"
	"github.com/giantswarm/inago/fleet"
	"github.com/giantswarm/inago/task"
	"github.com/giantswarm/inago/validator"
)

// Config provides all necessary and injectable configurations for a new
// controller.
type Config struct {
	// Dependencies.

	Fleet fleet.Fleet

	TaskService task.Service

	// Settings.

	// WaitCount represents the amount of times a desired status is required to
	// be seen to interpret it as final. E.g. when WaitCount is 3 and you start a
	// group, all statuses of units of that group need to be seen as "running" 3
	// times in a row.
	WaitCount int

	// WaitSleep represents the time to sleep between status-check cycles.
	WaitSleep time.Duration

	// WaitTimeout represents the maximum time to wait to reach a certain
	// status. When the desired status was not reached within the given period of
	// time, the wait ends.
	WaitTimeout time.Duration
}

// DefaultConfig provides a set of configurations with default values by best
// effort.
func DefaultConfig() Config {
	newFleetConfig := fleet.DefaultConfig()
	newFleet, err := fleet.NewFleet(newFleetConfig)
	if err != nil {
		panic(err)
	}

	newTaskServiceConfig := task.DefaultConfig()
	newTaskService := task.NewTaskService(newTaskServiceConfig)

	newConfig := Config{
		Fleet:       newFleet,
		TaskService: newTaskService,
		WaitCount:   3,
		WaitSleep:   1 * time.Second,
		WaitTimeout: 5 * time.Minute,
	}

	return newConfig
}

// NewController creates a new Controller that is configured with the given
// settings.
//
//   newConfig := controller.DefaultConfig()
//   newConfig.Fleet = myCustomFleetClient
//   newController := controller.NewController(newConfig)
//
func NewController(config Config) api.Controller {
	newController := controller{
		Config: config,
	}

	return &newController
}

type controller struct {
	Config
}

func (c controller) Submit(req api.Request) (*task.Task, error) {
	ok, err := validator.ValidateRequest(req)
	if !ok {
		return nil, maskAny(err)
	}

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

		closer := make(chan struct{})
		err = c.WaitForStatus(req, StatusStopped, closer)
		if err != nil {
			return maskAny(err)
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

func (c controller) Start(req api.Request) (*task.Task, error) {
	action := func() error {
		unitStatusList, err := c.groupStatusWithValidate(req)
		if err != nil {
			return maskAny(err)
		}

		for _, unitStatus := range unitStatusList {
			err := c.Fleet.Start(unitStatus.Name)
			if err != nil {
				return maskAny(err)
			}
		}

		closer := make(chan struct{})
		err = c.WaitForStatus(req, StatusRunning, closer)
		if err != nil {
			return maskAny(err)
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

func (c controller) Stop(req api.Request) (*task.Task, error) {
	action := func() error {
		unitStatusList, err := c.groupStatusWithValidate(req)
		if err != nil {
			return maskAny(err)
		}

		for _, unitStatus := range unitStatusList {
			err := c.Fleet.Stop(unitStatus.Name)
			if err != nil {
				return maskAny(err)
			}
		}

		closer := make(chan struct{})
		err = c.WaitForStatus(req, StatusStopped, closer)
		if err != nil {
			return maskAny(err)
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

func (c controller) Destroy(req api.Request) (*task.Task, error) {
	action := func() error {
		unitStatusList, err := c.groupStatusWithValidate(req)
		if err != nil {
			return maskAny(err)
		}

		for _, unitStatus := range unitStatusList {
			err := c.Fleet.Destroy(unitStatus.Name)
			if err != nil {
				return maskAny(err)
			}
		}

		closer := make(chan struct{})
		err = c.WaitForStatus(req, StatusNotFound, closer)
		if err != nil {
			return maskAny(err)
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

func (c controller) GetStatus(req api.Request) ([]fleet.UnitStatus, error) {
	status, err := c.groupStatusWithValidate(req)
	return status, maskAny(err)
}

func (c controller) WaitForStatus(req api.Request, desired api.Status, closer <-chan struct{}) error {
	fail := make(chan error)
	done := make(chan struct{})

	go func() {
		// count describes the count of how often the desired aggregated status was
		// seen.
		count := 0

	L1:
		for {
			unitStatusList, err := c.groupStatus(req)
			if IsUnitNotFound(err) && desired == StatusNotFound {
				goto C1
			} else if err != nil {
				fail <- maskAny(err)
				return
			}

			for _, us := range unitStatusList {
				for _, ms := range us.Machine {
					aggregated, err := AggregateStatus(us.Current, us.Desired, ms.SystemdActive, ms.SystemdSub)
					if err != nil {
						fail <- maskAny(err)
						return
					}

					if aggregated != desired {
						// Whenever the aggregated status does not match the desired
						// status, we reset the counter.
						count = 0
						time.Sleep(c.WaitSleep)
						continue L1
					}
				}
			}

		C1:
			count++
			if count == c.WaitCount {
				// In case the desired state was seen 3 times in a row, we assume we
				// finally reached the state we want to have.
				break
			}
			time.Sleep(c.WaitSleep)
		}

		done <- struct{}{}
	}()

	select {
	case err := <-fail:
		return maskAny(err)
	case <-done:
		return nil
	case <-closer:
		return nil
	case <-time.After(c.WaitTimeout):
		return maskAny(waitTimeoutReachedError)
	}
}

func (c controller) WaitForTask(taskID string, closer <-chan struct{}) (*task.Task, error) {
	taskObject, err := c.TaskService.WaitForFinalStatus(taskID, closer)
	return taskObject, maskAny(err)
}

// groupStatus fetches the group status using information provided
// by req. Note that this methods throws a unitNotFoundError in case no unit
// can be found.
func (c controller) groupStatus(req api.Request) ([]fleet.UnitStatus, error) {
	unitStatusList, err := c.Fleet.GetStatusWithMatcher(matchesGroupSlices(req))
	if fleet.IsUnitNotFound(err) {
		// This happens when no unit is found.
		return nil, maskAny(unitNotFoundError)
	} else if err != nil {
		return nil, maskAny(err)
	}

	// TODO retry operations

	return unitStatusList, nil
}

// groupStatusWithValidate fetches the group status using information provided
// by req. Note that this methods throws a unitNotFoundError in case no unit
// can be found, and a unitSliceNotFoundError in case at least one unit cannot
// be found.
func (c controller) groupStatusWithValidate(req api.Request) ([]fleet.UnitStatus, error) {
	unitStatusList, err := c.groupStatus(req)
	if err != nil {
		return nil, maskAny(err)
	}

	err = validateUnitStatusWithRequest(unitStatusList, req)
	if err != nil {
		return nil, maskAny(err)
	}

	// TODO retry operations

	return unitStatusList, nil
}

func validateUnitStatusWithRequest(unitStatusList []fleet.UnitStatus, req api.Request) error {
	for _, sliceID := range req.SliceIDs {
		if !containsUnitStatusSliceID(unitStatusList, sliceID) {
			// This happens when at least one of the units is not found.
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
func matchesGroupSlices(request api.Request) func(string) bool {
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
