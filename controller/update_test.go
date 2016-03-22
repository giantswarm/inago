package controller

import (
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/giantswarm/inago/task"
)

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
		function func(context.Context, Request) (*task.Task, error)
		ctx      context.Context
		req      Request
		err      error
	}{
		// Test a function that does nothing.
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
			ctx: context.Background(),
			req: Request{},
			err: nil,
		},
	}

	for i, test := range tests {
		err := testController.executeTaskAction(
			test.function,
			test.ctx,
			test.req,
		)

		if err != test.err {
			t.Logf("%v: method returned error '%v', which does not match test error '%v'", i, err, test.err)
			t.Fail()
		}
	}
}
