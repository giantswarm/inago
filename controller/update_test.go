package controller

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"

	"github.com/giantswarm/inago/fleet"
	"github.com/giantswarm/inago/logging"
	"github.com/giantswarm/inago/task"
)

// Return a controller for testing the updates with.
func getTestController() (controller, *fleetMock) {
	newFleetMock := newFleetMock(defaultFleetMockConfig())

	newTaskServiceConfig := task.DefaultConfig()
	newTaskServiceConfig.WaitSleep = 1 * time.Millisecond

	newTaskService := task.NewTaskService(newTaskServiceConfig)

	newLoggingConfig := logging.DefaultConfig()
	newLoggingConfig.LogLevel = "DEBUG"
	newLoggingConfig.Color = true
	newLogger := logging.NewLogger(newLoggingConfig)

	newControllerConfig := DefaultConfig()
	newControllerConfig.Fleet = newFleetMock
	newControllerConfig.TaskService = newTaskService
	newControllerConfig.Logger = newLogger
	newControllerConfig.WaitCount = 1
	newControllerConfig.WaitSleep = 1 * time.Millisecond
	newControllerConfig.WaitTimeout = 30 * time.Millisecond

	newController := controller{newControllerConfig}

	return newController, newFleetMock
}

// TestExecuteTaskAction tests the executeTaskAction method.
func TestExecuteTaskAction(t *testing.T) {
	testController, _ := getTestController()

	var tests = []struct {
		function   func(context.Context, Request) (*task.Task, error)
		ctx        context.Context
		req        Request
		errMatcher func(err error) bool
	}{
		// Test a task that does nothing.
		{
			function: func(ctx context.Context, req Request) (*task.Task, error) {
				taskObject, _ := testController.TaskService.Create(
					ctx,
					func(ctx context.Context) error {
						time.Sleep(5 * time.Millisecond)
						return nil
					},
				)

				return taskObject, nil
			},
			ctx:        context.Background(),
			req:        Request{},
			errMatcher: nil,
		},
	}

	for i, test := range tests {
		err := testController.executeTaskAction(
			test.function,
			test.ctx,
			test.req,
		)

		if err != nil && test.errMatcher == nil {
			t.Logf("%v: method return error '%v', when it should not return any errors", i, err)
			t.Fail()
		}

		if err != nil && test.errMatcher != nil && !test.errMatcher(err) {
			t.Logf("%v: method returned unexpected error '%v'", i, err)
			t.Fail()
		}
	}
}

// TestGetNumRunningSlices tests the getNumRunningSlices method.
func TestGetNumRunningSlices(t *testing.T) {
	var tests = []struct {
		fleetMockSetUp func(*fleetMock)
		req            Request
		numSlices      int
		errMatcher     func(err error) bool
	}{
		// Test that zero slices are returned if no unit statuses are returned by fleet,
		// and we don't ask for a group.
		{
			fleetMockSetUp: func(f *fleetMock) {
				f.On("GetStatusWithMatcher", mock.AnythingOfType("func(string) bool")).Return(
					[]fleet.UnitStatus{},
					nil,
				)
			},
			req:        Request{},
			numSlices:  0,
			errMatcher: nil,
		},
		// Test that IsUnitNotFound is returned if we give a group name,
		// but no unit statuses are returned by fleet.
		{
			fleetMockSetUp: func(f *fleetMock) {
				f.On("GetStatusWithMatcher", mock.AnythingOfType("func(string) bool")).Return(
					[]fleet.UnitStatus{},
					nil,
				)
			},
			req: Request{
				RequestConfig: RequestConfig{
					Group: "some group",
				},
			},
			numSlices:  0,
			errMatcher: IsUnitNotFound,
		},
		// Test that 1 slice is found if we give a group name,
		// and one unit status is returned by fleet.
		{
			fleetMockSetUp: func(f *fleetMock) {
				f.On("GetStatusWithMatcher", mock.AnythingOfType("func(string) bool")).Return(
					[]fleet.UnitStatus{
						fleet.UnitStatus{
							Name: "kubernetes-api-server",
						},
					},
					nil,
				)
			},
			req: Request{
				RequestConfig: RequestConfig{
					Group: "kubernetes",
				},
			},
			numSlices:  1,
			errMatcher: nil,
		},
	}

	for i, test := range tests {
		testController, fleetMock := getTestController()

		if test.fleetMockSetUp != nil {
			test.fleetMockSetUp(fleetMock)
		}

		numSlices, err := testController.getNumRunningSlices(context.Background(), test.req)

		if err != nil && test.errMatcher == nil {
			t.Logf("%v: method returned unexpected error '%v'", i, err)
			t.Fail()
		}

		if err != nil && test.errMatcher != nil && !test.errMatcher(err) {
			t.Logf("%v: method returned error '%v', which did not match expected error", i, err)
			t.Fail()
		}

		if numSlices != test.numSlices {
			t.Logf("%v: returned number of slices '%v' did not match expected: '%v'", i, numSlices, test.numSlices)
			t.Fail()
		}
	}
}

// TestIsGroupRemovalAllowed tests the isGroupRemovalAllowed method.
func TestIsGroupRemovalAllowed(t *testing.T) {
	var tests = []struct {
		fleetMockSetUp      func(*fleetMock)
		req                 Request
		minAlive            int
		groupRemovalAllowed bool
		errMatcher          func(err error) bool
	}{
		// Test group removal is not allowed if we ask to keep 0 alive,
		// and fleet has no units in it.
		{
			fleetMockSetUp: func(f *fleetMock) {
				f.On("GetStatusWithMatcher", mock.AnythingOfType("func(string) bool")).Return(
					[]fleet.UnitStatus{},
					nil,
				)
			},
			req:                 Request{},
			minAlive:            0,
			groupRemovalAllowed: false,
			errMatcher:          nil,
		},
	}

	for i, test := range tests {
		testController, fleetMock := getTestController()

		if test.fleetMockSetUp != nil {
			test.fleetMockSetUp(fleetMock)
		}

		groupRemovalAllowed, err := testController.isGroupRemovalAllowed(context.Background(), test.req, test.minAlive)

		if err != nil && test.errMatcher == nil {
			t.Logf("%v: method returned unexpected error '%v'", i, err)
			t.Fail()
		}

		if err != nil && test.errMatcher != nil && !test.errMatcher(err) {
			t.Logf("%v: method returned error '%v', which did not match expected error", i, err)
			t.Fail()
		}

		if groupRemovalAllowed != test.groupRemovalAllowed {
			t.Logf("%v: returned bool '%v' did not match expected: '%v'", i, groupRemovalAllowed, test.groupRemovalAllowed)
			t.Fail()
		}
	}
}

// TestUpdateWithStrategy tests the UpdateWithStrategy method.
func TestUpdateWithStrategy(t *testing.T) {
	tests := []struct {
		fleetMockSetUp func(*fleetMock)
		ctx            context.Context
		req            Request
		opts           UpdateOptions
		errMatcher     func(error) bool
	}{
		// Test that IsWaitTimeoutReached is returned if we attempt to
		// update given nil arguments.
		{
			fleetMockSetUp: nil,
			ctx:            context.Background(),
			req:            Request{},
			opts:           UpdateOptions{},
			errMatcher:     IsWaitTimeoutReached,
		},
	}

	for i, test := range tests {
		testController, fleetMock := getTestController()

		if test.fleetMockSetUp != nil {
			test.fleetMockSetUp(fleetMock)
		}

		err := testController.UpdateWithStrategy(test.ctx, test.req, test.opts)

		if err != nil && test.errMatcher == nil {
			t.Logf("%v: method returned unexpected error '%v'", i, err)
			t.Fail()
		}

		if err != nil && test.errMatcher != nil && !test.errMatcher(err) {
			t.Logf("%v: method returned error '%v', which did not match expected error", i, err)
			t.Fail()
		}
	}
}
