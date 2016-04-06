// Package controller implements a controller client providing basic operations against a
// controller endpoint through controller's HTTP API. Higher level scheduling and
// management should be built on top of that.
package controller

import (
	"strings"
	"time"

	"github.com/coreos/fleet/unit"
	"github.com/juju/errgo"
	"golang.org/x/net/context"

	"github.com/giantswarm/inago/common"
	"github.com/giantswarm/inago/fleet"
	"github.com/giantswarm/inago/logging"
	"github.com/giantswarm/inago/task"
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

	// Logger provides an initialised logger.
	Logger logging.Logger
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
		Logger:      logging.NewLogger(logging.DefaultConfig()),
	}

	return newConfig
}

// Controller defines the interface a controller needs to implement to provide
// operations for groups of unit files against a fleet cluster.
type Controller interface {
	ExtendWithExistingSliceIDs(req Request) (Request, error)

	// GroupNeedsUpdate checks if the given group should be updated or not. To
	// make a decision the unit content of each unit of each slice is compared
	// using its unit hash. As soon as one unit hash differs, or a unit cannot be
	// found, Inago assumes the whole group slice to be "dirty" and returns true
	// having the group slices removed from the given req that are up to date,
	// otherwise false, leaving the req as it is.
	GroupNeedsUpdate(ctx context.Context, req Request) (Request, bool, error)

	// Submit schedules a group on the configured fleet cluster. This is done by
	// setting the state of the units in the group to loaded.
	// If req.DesiredSlices is positive, new random (non conflicting) SliceIDs will be generated.
	// Otherwise the given req.SliceIDs will be used. Only one of those options can be used.
	Submit(ctx context.Context, req Request) (*task.Task, error)

	// Start starts a group on the configured fleet cluster. This is done by
	// setting the state of the units in the group to launched.
	Start(ctx context.Context, req Request) (*task.Task, error)

	// Stop stops a group on the configured fleet cluster. This is done by
	// setting the state of the units in the group to loaded.
	Stop(ctx context.Context, req Request) (*task.Task, error)

	// Destroy delets a group on the configured fleet cluster. This is done by
	// setting the state of the units in the group to inactive.
	Destroy(ctx context.Context, req Request) (*task.Task, error)

	// GetStatus fetches the current status of a group. If the unit cannot be
	// found, an error that you can identify using IsUnitNotFound is returned.
	GetStatus(ctx context.Context, req Request) ([]fleet.UnitStatus, error)

	// WaitForStatus waits for a group to reach the given status.
	WaitForStatus(ctx context.Context, req Request, closer <-chan struct{}, desiredStatuses ...Status) error

	// WaitForTask waits for the given task to reach a final status. Once the
	// given task has reached the final status, the final task representation is
	// returned.
	WaitForTask(ctx context.Context, taskID string, closer <-chan struct{}) (*task.Task, error)

	// Update updates the given group on best effort with respect to the given
	// opts. The given req identifies the group to update. The given options
	// define the strategy used to update the given group. See also
	// UpdateOptions.
	Update(ctx context.Context, req Request, opts UpdateOptions) (*task.Task, error)
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

func (c controller) GroupNeedsUpdate(ctx context.Context, req Request) (Request, bool, error) {
	c.Config.Logger.Debug(ctx, "controller: checking if group '%v' needs update", req.Group)

	// This becomes the new filtered list with all dirty slice IDs that needs to
	// be updated.
	var newSliceIDs []string

	c.Config.Logger.Debug(ctx, "controller: checking group status")
	usl, err := c.groupStatus(ctx, req)
	if err != nil {
		return Request{}, false, maskAny(err)
	}
	c.Config.Logger.Debug(ctx, "controller: checking unit hash info")
	uhis, err := groupUnitHashInfos(usl)
	if err != nil {
		return Request{}, false, maskAny(err)
	}
	c.Config.Logger.Debug(ctx, "controller: checking slice IDs")
	for _, u := range req.Units {
		unitFile, err := unit.NewUnitFile(u.Content)
		if err != nil {
			return Request{}, false, maskAny(err)
		}
		hash := unitFile.Hash().String()

		for _, uhi := range uhis {
			if common.UnitBase(u.Name) != uhi.Base {
				continue
			}
			if hash == uhi.Hash {
				continue
			}
			if contains(newSliceIDs, uhi.SliceID) {
				// We already tracked this ID. Go ahead.
				continue
			}

			newSliceIDs = append(newSliceIDs, uhi.SliceID)
		}
	}
	c.Config.Logger.Debug(ctx, "controller: finished checking slice IDs")

	if newSliceIDs == nil {
		return req, false, nil
	}
	req.SliceIDs = newSliceIDs

	c.Config.Logger.Debug(ctx, "controller: new slice IDs: %v", newSliceIDs)

	return req, true, nil
}

func (c controller) Submit(ctx context.Context, req Request) (*task.Task, error) {
	c.Config.Logger.Debug(ctx, "controller: handling submit")
	if ok, err := ValidateSubmitRequest(req); !ok {
		return nil, errgo.Cause(err)
	}
	action := func(ctx context.Context) error {
		var err error
		if req.DesiredSlices > 0 {
			req, err = c.ExtendWithRandomSliceIDs(ctx, req)
			if err != nil {
				return err
			}
		}

		req, err = req.ExtendSlices()
		if err != nil {
			return maskAny(err)
		}

		c.Config.Logger.Debug(ctx, "action: submitting units")
		for _, unit := range req.Units {
			err := c.Fleet.Submit(ctx, unit.Name, unit.Content)
			if err != nil {
				return maskAny(err)
			}
		}

		c.Config.Logger.Debug(ctx, "action: waiting for status of submitted units")
		closer := make(chan struct{})
		err = c.WaitForStatus(ctx, req, closer, StatusStopped)
		if err != nil {
			return maskAny(err)
		}

		// TODO retry operations

		return nil
	}
	taskObject, err := c.TaskService.Create(ctx, action)
	if err != nil {
		return nil, maskAny(err)
	}

	return taskObject, nil
}

func (c controller) Start(ctx context.Context, req Request) (*task.Task, error) {
	c.Config.Logger.Debug(ctx, "controller: handling start")

	action := func(ctx context.Context) error {
		c.Config.Logger.Debug(ctx, "action: fetching unit status list")
		unitStatusList, err := c.groupStatusWithValidate(ctx, req)
		if err != nil {
			return maskAny(err)
		}

		c.Config.Logger.Debug(ctx, "action: starting units")
		for _, unitStatus := range unitStatusList {
			err := c.Fleet.Start(ctx, unitStatus.Name)
			if err != nil {
				return maskAny(err)
			}
		}

		c.Config.Logger.Debug(ctx, "action: waiting for status of started units")
		closer := make(chan struct{})
		err = c.WaitForStatus(ctx, req, closer, StatusRunning)
		if err != nil {
			return maskAny(err)
		}

		// TODO retry operations

		return nil
	}

	taskObject, err := c.TaskService.Create(ctx, action)
	if err != nil {
		return nil, maskAny(err)
	}

	return taskObject, nil
}

func (c controller) Stop(ctx context.Context, req Request) (*task.Task, error) {
	c.Config.Logger.Debug(ctx, "controller: handling stop")

	action := func(ctx context.Context) error {
		unitStatusList, err := c.groupStatusWithValidate(ctx, req)
		if err != nil {
			return maskAny(err)
		}

		for _, unitStatus := range unitStatusList {
			err := c.Fleet.Stop(ctx, unitStatus.Name)
			if err != nil {
				return maskAny(err)
			}
		}

		closer := make(chan struct{})
		err = c.WaitForStatus(ctx, req, closer, StatusStopped, StatusFailed)
		if err != nil {
			return maskAny(err)
		}

		// TODO retry operations

		return nil
	}

	taskObject, err := c.TaskService.Create(ctx, action)
	if err != nil {
		return nil, maskAny(err)
	}

	return taskObject, nil
}

func (c controller) Destroy(ctx context.Context, req Request) (*task.Task, error) {
	c.Config.Logger.Debug(ctx, "controller: handling destroy")

	action := func(ctx context.Context) error {
		unitStatusList, err := c.groupStatusWithValidate(ctx, req)
		if err != nil {
			return maskAny(err)
		}

		for _, unitStatus := range unitStatusList {
			err := c.Fleet.Destroy(ctx, unitStatus.Name)
			if err != nil {
				return maskAny(err)
			}
		}

		closer := make(chan struct{})
		err = c.WaitForStatus(ctx, req, closer, StatusNotFound)
		if err != nil {
			return maskAny(err)
		}

		// TODO retry operations

		return nil
	}

	taskObject, err := c.TaskService.Create(ctx, action)
	if err != nil {
		return nil, maskAny(err)
	}

	return taskObject, nil
}

func (c controller) Update(ctx context.Context, req Request, opts UpdateOptions) (*task.Task, error) {
	c.Config.Logger.Debug(ctx, "controller: handling update for group: %v", req.Group)

	numRunning, err := c.getNumRunningSlices(ctx, req)
	if err != nil {
		return nil, maskAny(err)
	}
	c.Config.Logger.Debug(
		ctx, "controller: running: %v, growth: %v, alive: %v, ready: %v",
		numRunning, opts.MaxGrowth, opts.MinAlive, opts.ReadySecs,
	)
	updateAllowedRules := []struct {
		// The human readable error message for this rule
		message string
		// broken is true when the update should not be allowed
		broken bool
	}{
		{
			message: "maximum units to create during update must be positive, or zero",
			broken:  opts.MaxGrowth < 0,
		},
		{
			message: "minimum alive units must be positive, or zero",
			broken:  opts.MinAlive < 0,
		},
		{
			message: "time between creating groups must be positive, or zero",
			broken:  opts.ReadySecs < 0,
		},
		{
			message: "cannot have minimum alive units greater than current number of units",
			broken:  opts.MinAlive > numRunning,
		},
		{
			message: "to keep all current units alive, max growth must be greater than 0",
			broken:  opts.MinAlive == numRunning && opts.MaxGrowth < 1,
		},
	}
	for _, rule := range updateAllowedRules {
		if rule.broken {
			return nil, maskAnyf(updateNotAllowedError, rule.message)
		}
	}

	action := func(ctx context.Context) error {
		req, ok, err := c.GroupNeedsUpdate(ctx, req)
		if err != nil {
			return maskAny(err)
		}

		if !ok {
			// Group does not need to be updated. Do nothing.
			c.Config.Logger.Debug(ctx, "controller: unit does not need updating")
			return maskAny(unitsAlreadyUpToDate)
		}

		err = c.UpdateWithStrategy(ctx, req, opts)
		if err != nil {
			c.Config.Logger.Error(ctx, "controller: error encountered updating: %v", err)
			return maskAny(err)
		}

		// TODO retry operations

		return nil
	}

	taskObject, err := c.TaskService.Create(ctx, action)
	if err != nil {
		c.Config.Logger.Error(ctx, "controller: Could not create update task: %v", err)
		return nil, maskAny(err)
	}

	c.Config.Logger.Debug(ctx, "controller: created task object: %#v", taskObject)

	return taskObject, nil
}

func (c controller) GetStatus(ctx context.Context, req Request) ([]fleet.UnitStatus, error) {
	c.Config.Logger.Debug(ctx, "controller: handling getting status")

	status, err := c.groupStatusWithValidate(ctx, req)
	return status, maskAny(err)
}

func (c controller) WaitForStatus(ctx context.Context, req Request, closer <-chan struct{}, desiredStatuses ...Status) error {
	c.Config.Logger.Debug(ctx, "controller: handling waiting for status")

	if len(desiredStatuses) == 0 {
		return maskAny(invalidArgumentError)
	}

	fail := make(chan error)
	done := make(chan struct{})

	go func() {
		// count describes the count of how often one of the desired aggregated statuses was
		// seen.
		count := 0

	L1:
		for {
			c.Config.Logger.Debug(ctx, "controller: fetching group status")

			unitStatusList, err := c.groupStatus(ctx, req)
			for _, desiredStatus := range desiredStatuses {
				if IsUnitNotFound(err) && desiredStatus == StatusNotFound {
					goto C1
				} else if err != nil {
					fail <- maskAny(err)
					return
				}
			}

			c.Config.Logger.Debug(ctx, "controller: checking units have desired statuses: %v", desiredStatuses)
			for _, us := range unitStatusList {
				c.Config.Logger.Debug(ctx, "controller: unit status: %#v", us)

				aggregator := Aggregator{
					Logger: c.Config.Logger,
				}
				ok, err := aggregator.UnitHasStatus(us, desiredStatuses...)
				if err != nil {
					fail <- maskAny(err)
					return
				}
				if !ok {
					c.Config.Logger.Debug(ctx, "controller: unit %v does not have desired statuses: %v", us, desiredStatuses)
					// Whenever the aggregated status does not match the desired
					// statuses, we reset the counter.
					count = 0
					time.Sleep(c.WaitSleep)
					continue L1
				}
			}

		C1:
			c.Config.Logger.Debug(ctx, "controller: group has desired statuses: %v", desiredStatuses)
			count++
			if count == c.WaitCount {
				// In case the desired statuses were seen 3 times in a row, we assume we
				// finally reached the status we want to have.
				c.Config.Logger.Debug(ctx, "controller: group has reached count (%v) of desired statuses: %v", c.WaitCount, desiredStatuses)
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

func (c controller) WaitForTask(ctx context.Context, taskID string, closer <-chan struct{}) (*task.Task, error) {
	c.Config.Logger.Debug(ctx, "controller: handling waiting for task")

	taskObject, err := c.TaskService.WaitForFinalStatus(ctx, taskID, closer)
	if err != nil {
		c.Config.Logger.Error(ctx, "controller: error occurred waiting for task: %#v", err)
	}
	return taskObject, maskAny(err)
}

// groupStatus fetches the group status using information provided
// by req. Note that this methods throws a unitNotFoundError in case no unit
// can be found.
func (c controller) groupStatus(ctx context.Context, req Request) ([]fleet.UnitStatus, error) {
	c.Config.Logger.Debug(ctx, "controller: fetching group status from fleet")

	unitStatusList, err := c.Fleet.GetStatusWithMatcher(matchesGroupSlices(req))
	if fleet.IsUnitNotFound(err) {
		// This happens when no unit is found.
		return nil, maskAny(unitNotFoundError)
	} else if err != nil {
		return nil, maskAny(err)
	}
	c.Config.Logger.Debug(ctx, "controller: received unit status list: %#v", unitStatusList)

	// TODO retry operations

	return unitStatusList, nil
}

// groupStatusWithValidate fetches the group status using information provided
// by req. Note that this methods throws a unitNotFoundError in case no unit
// can be found, and a unitSliceNotFoundError in case at least one unit cannot
// be found.
func (c controller) groupStatusWithValidate(ctx context.Context, req Request) ([]fleet.UnitStatus, error) {
	unitStatusList, err := c.groupStatus(ctx, req)
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

func validateUnitStatusWithRequest(unitStatusList []fleet.UnitStatus, req Request) error {
	for _, sliceID := range req.SliceIDs {
		ok, err := containsUnitStatusSliceID(unitStatusList, sliceID)
		if err != nil {
			return maskAny(err)
		}
		if !ok {
			// This happens when at least one of the units is not found.
			return maskAnyf(unitSliceNotFoundError, "slice ID '%s'", sliceID)
		}
	}

	return nil
}

func containsUnitStatusSliceID(unitStatusList []fleet.UnitStatus, sliceID string) (bool, error) {
	for _, us := range unitStatusList {
		ID, err := common.SliceID(us.Name)
		if err != nil {
			return false, maskAny(err)
		}
		if ID == sliceID {
			return true, nil
		}
	}

	return false, nil
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
	return func(unitName string) bool {
		if !strings.HasPrefix(unitName, request.Group) {
			return false
		}

		for _, sliceID := range request.SliceIDs {
			// TODO fix extension
			if strings.HasSuffix(unitName, "@"+sliceID+".service") {
				return true
			}
		}

		return false
	}
}

func matchesUnitBase(request Request) func(string) bool {
	// If only the group name is of interest, return shorter version
	if request.Units == nil || len(request.Units) == 0 {
		return func(name string) bool {
			return strings.HasPrefix(name, request.Group)
		}
	}

	// Normal version that matches on group prefix and slice ID suffix.
	return func(unitName string) bool {
		if !strings.HasPrefix(unitName, request.Group) {
			return false
		}

		for _, u := range request.Units {
			if common.UnitBase(u.Name) == common.UnitBase(unitName) {
				return true
			}
		}

		return false
	}
}
