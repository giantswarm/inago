package controller

import (
	"testing"
	"time"

	"github.com/juju/errgo"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"

	"github.com/giantswarm/inago/fleet"
	"github.com/giantswarm/inago/task"
)

var (
	testError = errgo.New("test error")
)

func IsTestError(err error) bool {
	return errgo.Cause(err) == testError
}

// Return a controller for testing the updates with.
func getTestController() (controller, *fleetMock) {
	newFleetMock := newFleetMock(defaultFleetMockConfig())

	newTaskServiceConfig := task.DefaultConfig()
	newTaskServiceConfig.WaitSleep = 1 * time.Millisecond

	newTaskService := task.NewTaskService(newTaskServiceConfig)

	newControllerConfig := DefaultConfig()
	newControllerConfig.Fleet = newFleetMock
	newControllerConfig.TaskService = newTaskService
	newControllerConfig.WaitCount = 1
	newControllerConfig.WaitSleep = 1 * time.Millisecond
	newControllerConfig.WaitTimeout = 3 * time.Millisecond

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
		// Test a task that returns an error.
		{
			function: func(ctx context.Context, req Request) (*task.Task, error) {
				taskObject, _ := testController.TaskService.Create(
					ctx,
					func(ctx context.Context) error {
						return testError
					},
				)

				return taskObject, nil
			},
			ctx:        context.Background(),
			req:        Request{},
			errMatcher: IsTestError,
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

		if err != nil && test.errMatcher != nil && test.errMatcher(err) {
			t.Logf("%v: method returned error '%v', which does not match error matcher '%v'", i, err, test.errMatcher)
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

		numSlices, err := testController.getNumRunningSlices(test.req)

		if err != nil && test.errMatcher == nil {
			t.Logf("%v: method returned unexpected error '%v'", i, err)
			t.Fail()
		}

		if err != nil && test.errMatcher != nil && test.errMatcher(err) {
			t.Logf("%v: method returned error '%v', which did not match expected error", i, err)
			t.Fail()
		}

		if numSlices != test.numSlices {
			t.Logf("%v: returned number of slices '%v' did not match expected: '%v'", i, numSlices, test.numSlices)
			t.Fail()
		}
	}
}
