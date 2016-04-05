package controller

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/giantswarm/inago/fleet"
	"github.com/giantswarm/inago/logging"
	"github.com/giantswarm/inago/task"
)

// Return a controller for testing the updates with.
func getTestController() (controller, *fleet.DummyFleet) {
	newTaskServiceConfig := task.DefaultConfig()
	newTaskServiceConfig.WaitSleep = 100 * time.Millisecond

	newTaskService := task.NewTaskService(newTaskServiceConfig)

	newLoggingConfig := logging.DefaultConfig()
	newLoggingConfig.LogLevel = "DEBUG"
	newLoggingConfig.Color = true
	newLogger := logging.NewLogger(newLoggingConfig)

	dummyFleetConfig := fleet.DefaultDummyConfig()
	dummyFleetConfig.Logger = newLogger
	dummyFleet := fleet.NewDummyFleet(dummyFleetConfig)

	newControllerConfig := DefaultConfig()
	newControllerConfig.Fleet = dummyFleet
	newControllerConfig.TaskService = newTaskService
	newControllerConfig.Logger = newLogger
	newControllerConfig.WaitCount = 1
	newControllerConfig.WaitSleep = 300 * time.Millisecond
	newControllerConfig.WaitTimeout = 5 * time.Second

	newController := controller{newControllerConfig}

	return newController, dummyFleet
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
		fleetSetUp func(f fleet.Fleet)
		req        Request
		numSlices  int
		errMatcher func(err error) bool
	}{
		// Test that zero slices are returned if no unit statuses are returned by fleet,
		// and we don't ask for a group.
		{
			req:        Request{},
			numSlices:  0,
			errMatcher: nil,
		},
		// Test that IsUnitNotFound is returned if we give a group name,
		// but no unit statuses are returned by fleet.
		{
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
			fleetSetUp: func(f fleet.Fleet) {
				f.Submit(
					context.Background(),
					"kubernetes-unit@1.service",
					"some content",
				)
				f.Start(
					context.Background(),
					"kubernetes-unit@1.service",
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
		testController, dummyFleet := getTestController()

		if test.fleetSetUp != nil {
			test.fleetSetUp(dummyFleet)
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
		fleetSetUp          func(f fleet.Fleet)
		req                 Request
		minAlive            int
		groupRemovalAllowed bool
		errMatcher          func(err error) bool
	}{
		// Test group removal is not allowed if we ask to keep 0 alive,
		// and fleet has no units in it.
		{
			req:                 Request{},
			minAlive:            0,
			groupRemovalAllowed: false,
			errMatcher:          nil,
		},
	}

	for i, test := range tests {
		testController, dummyFleet := getTestController()

		if test.fleetSetUp != nil {
			test.fleetSetUp(dummyFleet)
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

// TestUpdateCurrentSliceIDs tests the updateCurrentSliceIDs method.
func TestUpdateCurrentSliceIDs(t *testing.T) {
	tests := []struct {
		currentSliceIDs  []string
		modifiedSliceIDs []string
		newSliceIDs      []string
		output           []string
	}{
		// Test the case where everything is empty.
		{
			currentSliceIDs:  []string{},
			modifiedSliceIDs: []string{},
			newSliceIDs:      []string{},
			output:           []string{},
		},
		// Test the case where no slices are added.
		{
			currentSliceIDs:  []string{"jf8", "8fh", "hf9"},
			modifiedSliceIDs: []string{"8fh"},
			newSliceIDs:      []string{},
			output:           []string{"jf8", "hf9"},
		},
		// Test the case where no slices are modified.
		{
			currentSliceIDs:  []string{"hfh", "9he", "jm2"},
			modifiedSliceIDs: []string{},
			newSliceIDs:      []string{"ns2"},
			output:           []string{"hfh", "9he", "jm2", "ns2"},
		},
		// Test the case where no slices are modified or added.
		{
			currentSliceIDs:  []string{"hfh", "9he", "jm2"},
			modifiedSliceIDs: []string{},
			newSliceIDs:      []string{},
			output:           []string{"hfh", "9he", "jm2"},
		},
		// Test the usual case, of one slice being modified, and one slice being added.
		{
			currentSliceIDs:  []string{"lol", "kek"},
			modifiedSliceIDs: []string{"lol"},
			newSliceIDs:      []string{"02b"},
			output:           []string{"kek", "02b"},
		},
	}

	for _, test := range tests {
		testController, _ := getTestController()

		returnedOutput := testController.updateCurrentSliceIDs(
			context.Background(),
			test.currentSliceIDs,
			test.modifiedSliceIDs,
			test.newSliceIDs,
		)

		if !reflect.DeepEqual(test.output, returnedOutput) {
			t.Fatalf("Should return: %v, returned: %v", test.output, returnedOutput)
		}
	}
}

// TestUpdateWithStrategy tests the UpdateWithStrategy method.
func TestUpdateWithStrategy(t *testing.T) {
	tests := []struct {
		fleetSetUp func(fleet.Fleet)
		req        Request
		opts       UpdateOptions
		assertion  func(*testing.T, *fleet.DummyFleet, error)
	}{
		// Test a minimal update.
		{
			fleetSetUp: func(f fleet.Fleet) {
				unitName := "bluebird-unit@1.service"

				f.Submit(context.Background(), unitName, "some content")
				f.Start(context.Background(), unitName)
			},
			req: Request{
				RequestConfig: RequestConfig{
					Group:    "bluebird",
					SliceIDs: []string{"1"},
				},
				Units: []Unit{
					{
						Name:    "bluebird-unit@.service",
						Content: "some updated content",
					},
				},
			},
			opts: UpdateOptions{
				MaxGrowth: 1,
				MinAlive:  0,
			},
			assertion: func(t *testing.T, f *fleet.DummyFleet, e error) {
				if e != nil {
					t.Fatal("Error returned by update:", e)
				}

				unitStatusList, err := f.GetStatusWithMatcher(
					func(s string) bool {
						return strings.HasPrefix(s, "bluebird-unit@") && strings.HasSuffix(s, ".service")
					},
				)
				if err != nil {
					t.Fatal("Error returned getting statuses: ", err)
				}
				if len(unitStatusList) != 1 {
					t.Fatal("Incorrect number of units:", len(f.Units))
				}
				unitStatus := unitStatusList[0]
				if unitStatus.Name == "bluebird-unit@1.service" {
					t.Fatal("Unit has same name as original")
				}
				if unitStatus.Current != "launched" {
					t.Fatal("Unit has incorrect current status")
				}
				if unitStatus.Desired != "launched" {
					t.Fatal("Unit has incorrect desired status")
				}
			},
		},
		// Test an update of a group with two group slices.
		{
			fleetSetUp: func(f fleet.Fleet) {
				unitNameTemplate := "canary-unit@%v.service"

				for _, id := range []string{"lol", "kek"} {
					unitName := fmt.Sprintf(unitNameTemplate, id)

					f.Submit(context.Background(), unitName, "some content")
					f.Start(context.Background(), unitName)
				}
			},
			req: Request{
				RequestConfig: RequestConfig{
					Group:    "canary",
					SliceIDs: []string{"lol", "kek"},
				},
				Units: []Unit{
					Unit{
						Name:    "canary-unit@.service",
						Content: "some updated content",
					},
				},
			},
			opts: UpdateOptions{
				MaxGrowth: 1,
				MinAlive:  1,
			},
			assertion: func(t *testing.T, f *fleet.DummyFleet, e error) {
				if e != nil {
					t.Fatal("Error returned by update:", e)
				}

				unitStatusList, err := f.GetStatusWithMatcher(
					func(s string) bool {
						return strings.HasPrefix(s, "canary-unit@") && strings.HasSuffix(s, ".service")
					},
				)
				if err != nil {
					t.Fatal("Error returned getting statuses: ", err)
				}

				if len(unitStatusList) != 2 {
					t.Fatal("Incorrect number of units:", len(unitStatusList))
				}

				for _, unitStatus := range unitStatusList {
					if unitStatus.Current != "launched" {
						t.Fatal("Incorrect current status:", unitStatus.Current)
					}
					if unitStatus.Desired != "launched" {
						t.Fatal("Incorrect desired status:", unitStatus.Desired)
					}
					if unitStatus.Name == "canary-unit@lol.service" || unitStatus.Name == "canary-unit@kek.service" {
						t.Fatal("Previous unit name in use:", unitStatus.Name)
					}
				}
			},
		},
		// Test a reasonable update of a group with two group slices.
		{
			fleetSetUp: func(f fleet.Fleet) {
				unitNameTemplate := "sparrow-unit@%v.service"

				for _, id := range []string{"ded", "bef"} {
					unitName := fmt.Sprintf(unitNameTemplate, id)

					f.Submit(context.Background(), unitName, "some content")
					f.Start(context.Background(), unitName)
				}
			},
			req: Request{
				RequestConfig: RequestConfig{
					Group:    "sparrow",
					SliceIDs: []string{"ded", "bef"},
				},
				Units: []Unit{
					Unit{
						Name:    "sparrow-unit@.service",
						Content: "some updated content",
					},
				},
			},
			opts: UpdateOptions{
				MaxGrowth: 0,
				MinAlive:  1,
			},
			assertion: func(t *testing.T, f *fleet.DummyFleet, e error) {
				if e != nil {
					t.Fatal("Error returned by update:", e)
				}

				unitStatusList, err := f.GetStatusWithMatcher(
					func(s string) bool {
						return strings.HasPrefix(s, "sparrow-unit@") && strings.HasSuffix(s, ".service")
					},
				)
				if err != nil {
					t.Fatal("Error returned getting statuses: ", err)
				}

				if len(unitStatusList) != 2 {
					t.Fatal("Incorrect number of units:", len(unitStatusList))
				}

				for _, unitStatus := range unitStatusList {
					if unitStatus.Current != "launched" {
						t.Fatal("Incorrect current status:", unitStatus.Current)
					}
					if unitStatus.Desired != "launched" {
						t.Fatal("Incorrect desired status:", unitStatus.Desired)
					}
					if unitStatus.Name == "sparrow-unit@ded.service" || unitStatus.Name == "sparrow-unit@bef.service" {
						t.Fatal("Previous unit name in use:", unitStatus.Name)
					}
				}
			},
		},
	}

	for i, test := range tests {
		fmt.Println("Running test", i)

		testController, dummyFleet := getTestController()

		test.fleetSetUp(dummyFleet)
		err := testController.UpdateWithStrategy(context.Background(), test.req, test.opts)
		test.assertion(t, dummyFleet, err)
	}
}
