package controller

import (
	"testing"
	"time"

	"github.com/juju/errgo"
	"golang.org/x/net/context"

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
