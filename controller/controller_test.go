package controller

import (
	"net"
	"testing"
	"time"

	"github.com/juju/errgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	"github.com/giantswarm/inago/controller/api"
	"github.com/giantswarm/inago/fleet"
	"github.com/giantswarm/inago/task"
)

func Test_Request_ExtendSlices(t *testing.T) {
	testCases := []struct {
		Input    api.Request
		Expected api.Request
	}{
		// This test ensures that the request is not manipulated when no slice IDs
		// are given.
		{
			Input: api.Request{
				SliceIDs: []string{},
				Units: []api.Unit{
					{
						Name:    "unit@.service",
						Content: "some unit content",
					},
				},
			},
			Expected: api.Request{
				SliceIDs: []string{},
				Units: []api.Unit{
					{
						Name:    "unit@.service",
						Content: "some unit content",
					},
				},
			},
		},

		// This test ensures that the request is extended when one slice ID is
		// given.
		{
			Input: api.Request{
				SliceIDs: []string{"1"},
				Units: []api.Unit{
					{
						Name:    "unit@.service",
						Content: "some unit content",
					},
				},
			},
			Expected: api.Request{
				SliceIDs: []string{"1"},
				Units: []api.Unit{
					{
						Name:    "unit@1.service",
						Content: "some unit content",
					},
				},
			},
		},

		// This test ensures that the request is extended when multiple slice IDs
		// and multiple unit files are given.
		{
			Input: api.Request{
				SliceIDs: []string{"1", "2"},
				Units: []api.Unit{
					{
						Name:    "foo@.service",
						Content: "some foo content",
					},
					{
						Name:    "bar@.service",
						Content: "some bar content",
					},
				},
			},
			Expected: api.Request{
				SliceIDs: []string{"1", "2"},
				Units: []api.Unit{
					{
						Name:    "foo@1.service",
						Content: "some foo content",
					},
					{
						Name:    "bar@1.service",
						Content: "some bar content",
					},
					{
						Name:    "foo@2.service",
						Content: "some foo content",
					},
					{
						Name:    "bar@2.service",
						Content: "some bar content",
					},
				},
			},
		},

		// This test ensures that the request is extended when arbitrary slice IDs
		// are given.
		{
			Input: api.Request{
				SliceIDs: []string{"3", "5", "foo"},
				Units: []api.Unit{
					{
						Name:    "unit@.service",
						Content: "some unit content",
					},
				},
			},
			Expected: api.Request{
				SliceIDs: []string{"3", "5", "foo"},
				Units: []api.Unit{
					{
						Name:    "unit@3.service",
						Content: "some unit content",
					},
					{
						Name:    "unit@5.service",
						Content: "some unit content",
					},
					{
						Name:    "unit@foo.service",
						Content: "some unit content",
					},
				},
			},
		},

		// This test ensures we generate the correct units for group instances,
		// so groups that do not want slices.
		{
			Input: api.Request{
				SliceIDs: nil,
				Units: []api.Unit{
					{
						Name:    "single-1.service",
						Content: "some unit content",
					},
					{
						Name:    "single-2.service",
						Content: "some unit content",
					},
				},
			},
			Expected: api.Request{
				SliceIDs: nil,
				Units: []api.Unit{
					{
						Name:    "single-1.service",
						Content: "some unit content",
					},
					{
						Name:    "single-2.service",
						Content: "some unit content",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		output, err := testCase.Input.ExtendSlices()
		if err != nil {
			t.Fatalf("Request.ExtendSlices returned error: %#v", err)
		}

		if len(output.Units) != len(testCase.Expected.Units) {
			t.Fatalf("Number of units in expected output differed from received units: %d != %d", len(output.Units), len(testCase.Expected.Units))
		}

		for i, outputUnit := range output.Units {
			if outputUnit.Name != testCase.Expected.Units[i].Name {
				t.Fatalf("output unit name '%s' is not equal to expected unit name '%s'", outputUnit.Name, testCase.Expected.Units[i].Name)
			}
		}
	}
}

func Test_validateUnitStatusWithRequest(t *testing.T) {
	testCases := []struct {
		Error          error
		Request        api.Request
		UnitStatusList []fleet.UnitStatus
	}{
		// This test ensures that validating a unit status list against given slice
		// IDs works as expected.
		{
			Error: nil,
			Request: api.Request{
				SliceIDs: []string{"1", "2"},
			},
			UnitStatusList: []fleet.UnitStatus{
				{
					Name: "name@1.service",
				},
				{
					Name: "name@2.service",
				},
			},
		},

		// This test ensures that validating a unit status list against given slice
		// IDs throws an error in case there is a unit missing
		{
			Error: errgo.New("unit slice not found: slice ID '3'"),
			Request: api.Request{
				SliceIDs: []string{"1", "2", "3"},
			},
			UnitStatusList: []fleet.UnitStatus{
				{
					Name: "name@1.service",
				},
				{
					Name: "name@2.service",
				},
			},
		},
	}

	for _, testCase := range testCases {
		err := validateUnitStatusWithRequest(testCase.UnitStatusList, testCase.Request)
		if testCase.Error != nil && err.Error() != testCase.Error.Error() {
			t.Fatalf("validateUnitStatusWithRequest returned error: %#v", err)
		}
	}
}

func Test_matchesGroupSlices(t *testing.T) {
	testCases := []struct {
		InputUnitName string
		InputRequest  api.Request
		Output        bool
	}{
		{
			InputUnitName: "demo-main@1.service",
			InputRequest: api.Request{
				Group:    "demo",
				SliceIDs: []string{"1", "2"},
			},
			Output: true,
		},
		{
			InputUnitName: "demo-main@1.service",
			InputRequest: api.Request{
				Group:    "demo",
				SliceIDs: []string{"3"},
			},
			Output: false,
		},
		{
			InputUnitName: "other-main@1.service",
			InputRequest: api.Request{
				Group:    "demo",
				SliceIDs: []string{"1"},
			},
			Output: false,
		},
		{
			InputUnitName: "other-main@1.service",
			InputRequest: api.Request{
				Group:    "demo",
				SliceIDs: []string{"2"},
			},
			Output: false,
		},

		{
			InputUnitName: "demo-main.service",
			InputRequest: api.Request{
				Group:    "demo",
				SliceIDs: nil,
			},
			Output: true,
		},

		{
			InputUnitName: "other-main.service",
			InputRequest: api.Request{
				Group:    "demo",
				SliceIDs: []string{"1", "2"},
			},
			Output: false,
		},
		{
			InputUnitName: "other-main.service",
			InputRequest: api.Request{
				Group:    "demo",
				SliceIDs: nil,
			},
			Output: false,
		},
	}

	for id, test := range testCases {
		matcher := matchesGroupSlices(test.InputRequest)
		result := matcher(test.InputUnitName)

		if result != test.Output {
			t.Errorf("TestCase %d: Failed to match: Input '%s', expected %v, got %v", id, test.InputUnitName, test.Output, result)
		}
	}
}

// givenController returns a controller where the fleet backend is replaced
// with a mock.
func givenController() (api.Controller, *fleetMock) {
	newFleetMockConfig := defaultFleetMockConfig()
	newController, newFleetMock := givenControllerWithConfig(newFleetMockConfig)

	return newController, newFleetMock
}

// givenController returns a controller where the fleet backend is replaced
// with a mock.
func givenControllerWithConfig(fmc fleetMockConfig) (api.Controller, *fleetMock) {
	newFleetMock := newFleetMock(fmc)

	newTaskServiceConfig := task.DefaultConfig()
	newTaskService := task.NewTaskService(newTaskServiceConfig)

	newControllerConfig := DefaultConfig()
	newControllerConfig.Fleet = newFleetMock
	newControllerConfig.TaskService = newTaskService
	newControllerConfig.WaitCount = 1
	newControllerConfig.WaitSleep = 10 * time.Millisecond
	newController := NewController(newControllerConfig)

	return newController, newFleetMock
}

func TestController_Submit_Error(t *testing.T) {
	RegisterTestingT(t)

	controller, fleetMock := givenController()

	// Execute
	req := api.Request{
		Group:    "single",
		SliceIDs: nil,
		Units:    []api.Unit{}, // Intentionally left empty!
	}

	task, err := controller.Submit(req)

	// Assert
	Expect(task).To(BeNil())
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(Equal("no units in group"))
	mock.AssertExpectationsForObjects(t, fleetMock.Mock)
}

func TestController_Submit(t *testing.T) {
	RegisterTestingT(t)

	// Mocks
	controller, fleetMock := givenController()
	fleetMock.On("GetStatusWithMatcher", mock.AnythingOfType("func(string) bool")).Return(
		[]fleet.UnitStatus{
			{
				Name: "test-main@1.service",
			},
		},
		nil,
	)
	fleetMock.On("Submit", "test-main@1.service", "content").Return(nil).Once()

	// Execute test
	req := api.Request{
		Group:    "test",
		SliceIDs: []string{"1"},
		Units: []api.Unit{
			{
				Name:    "test-main@1.service",
				Content: "content",
			},
		},
	}
	taskObject, err := controller.Submit(req)
	Expect(err).To(BeNil())

	_, err = controller.WaitForTask(taskObject.ID, nil)
	Expect(err).To(BeNil())

	// Assert
	mock.AssertExpectationsForObjects(t, fleetMock.Mock)
}

func TestController_Start(t *testing.T) {
	RegisterTestingT(t)

	// Mocks
	controller, fleetMock := givenController()
	fleetMock.On("GetStatusWithMatcher", mock.AnythingOfType("func(string) bool")).Return(
		[]fleet.UnitStatus{
			{
				Name: "test-main@1.service",
			},
			{
				Name: "test-sidekick@1.service",
			},
		},
		nil,
	)
	fleetMock.On("Start", "test-main@1.service").Return(nil).Once()
	fleetMock.On("Start", "test-sidekick@1.service").Return(nil).Once()

	// Execute test
	req := api.Request{
		Group:    "test",
		SliceIDs: []string{"1"},
	}
	taskObject, err := controller.Start(req)
	Expect(err).To(BeNil())

	_, err = controller.WaitForTask(taskObject.ID, nil)
	Expect(err).To(BeNil())

	// Assert
	mock.AssertExpectationsForObjects(t, fleetMock.Mock)
}

func TestController_Destroy(t *testing.T) {
	RegisterTestingT(t)

	// Mocks
	controller, fleetMock := givenController()
	fleetMock.On("GetStatusWithMatcher", mock.AnythingOfType("func(string) bool")).Return(
		[]fleet.UnitStatus{
			{
				Name: "test-main@1.service",
			},
			{
				Name: "test-sidekick@1.service",
			},
		},
		nil,
	)
	fleetMock.On("Destroy", "test-main@1.service").Return(nil).Once()
	fleetMock.On("Destroy", "test-sidekick@1.service").Return(nil).Once()

	// Execute test
	req := api.Request{
		Group:    "test",
		SliceIDs: []string{"1"},
	}
	taskObject, err := controller.Destroy(req)
	Expect(err).To(BeNil())

	_, err = controller.WaitForTask(taskObject.ID, nil)
	Expect(err).To(BeNil())

	// Assert
	mock.AssertExpectationsForObjects(t, fleetMock.Mock)
}

func TestController_Stop(t *testing.T) {
	RegisterTestingT(t)

	// Mocks
	controller, fleetMock := givenController()
	fleetMock.On("GetStatusWithMatcher", mock.AnythingOfType("func(string) bool")).Return(
		[]fleet.UnitStatus{
			{
				Name: "test-main@1.service",
			},
			{
				Name: "test-sidekick@1.service",
			},
		},
		nil,
	)
	fleetMock.On("Stop", "test-main@1.service").Return(nil).Once()
	fleetMock.On("Stop", "test-sidekick@1.service").Return(nil).Once()

	// Execute test
	req := api.Request{
		Group:    "test",
		SliceIDs: []string{"1"},
	}
	taskObject, err := controller.Stop(req)
	Expect(err).To(BeNil())

	_, err = controller.WaitForTask(taskObject.ID, nil)
	Expect(err).To(BeNil())

	// Assert
	mock.AssertExpectationsForObjects(t, fleetMock.Mock)
}

func TestController_Status_ErrorOnMismatchingSliceIDs(t *testing.T) {
	RegisterTestingT(t)

	// Mocks
	controller, fleetMock := givenController()
	fleetMock.On("GetStatusWithMatcher", mock.AnythingOfType("func(string) bool")).Return(
		[]fleet.UnitStatus{
			{
				Name: "test-main@1.service",
			},
		},
		nil,
	).Once()

	// Execute
	status, err := controller.GetStatus(api.Request{
		Group:    "test",
		SliceIDs: []string{"1", "2"},
	})

	// Assert
	Expect(IsUnitSliceNotFound(err)).To(Equal(true))
	Expect(status).To(BeEmpty())
	mock.AssertExpectationsForObjects(t, fleetMock.Mock)
}

// TestController_WaitForStatus_Success tests the normal behaviour of
// Controller.WaitForStatus to ensure it works as expected.
func TestController_WaitForStatus_Success(t *testing.T) {
	RegisterTestingT(t)

	// Mocks
	newFleetMockConfig := defaultFleetMockConfig()
	newFleetMockConfig.UseTestifyMock = false
	newFleetMockConfig.UseCustomMock = true
	newFleetMockConfig.FirstCustomMockStatus = []fleet.UnitStatus{
		{
			Current: "loaded",
			Desired: "loaded",
			Machine: []fleet.MachineStatus{
				{
					ID:            "test-id",
					IP:            net.ParseIP("10.0.0.101"),
					SystemdActive: "activating",
					SystemdSub:    "start-pre",
					UnitHash:      "test-hash",
				},
			},
			Name: "test-main@1.service",
		},
	}
	newFleetMockConfig.LastCustomMockStatus = []fleet.UnitStatus{
		{
			Current: "loaded",
			Desired: "loaded",
			Machine: []fleet.MachineStatus{
				{
					ID:            "test-id",
					IP:            net.ParseIP("10.0.0.101"),
					SystemdActive: "active",
					SystemdSub:    "running",
					UnitHash:      "test-hash",
				},
			},
			Name: "test-main@1.service",
		},
	}

	controller, fleetMock := givenControllerWithConfig(newFleetMockConfig)
	fleetMock.On("Start", "test-main@1.service").Return(nil).Once()

	// Execute test
	req := api.Request{
		Group:    "test",
		SliceIDs: []string{"1"},
		Units: []api.Unit{
			{
				Name:    "test-main@1.service",
				Content: "content",
			},
		},
	}
	taskObject, err := controller.Start(req)
	Expect(err).To(BeNil())
	Expect(task.HasFinalStatus(taskObject)).To(Not(BeTrue()))

	taskObject, err = controller.WaitForTask(taskObject.ID, nil)
	Expect(err).To(BeNil())
	Expect(task.HasSucceededStatus(taskObject)).To(BeTrue())

	// Assert
	mock.AssertExpectationsForObjects(t, fleetMock.Mock)
}

// TestController_WaitForStatus_Closer tests Controller.WaitForStatus to
// directly end waiting when the closer is used.
func TestController_WaitForStatus_Closer(t *testing.T) {
	RegisterTestingT(t)

	// Mocks
	controller, fleetMock := givenController()
	fleetMock.On("GetStatusWithMatcher", mock.AnythingOfType("func(string) bool")).Return(
		[]fleet.UnitStatus{
			{
				Name: "test-main@1.service",
			},
			{
				Name: "test-sidekick@1.service",
			},
		},
		nil,
	)

	// Execute test
	req := api.Request{
		Group:    "test",
		SliceIDs: []string{"1"},
		Units: []api.Unit{
			{
				Name:    "test-main@1.service",
				Content: "content",
			},
		},
	}
	desired := StatusRunning
	closer := make(chan struct{}, 1)
	closer <- struct{}{}

	err := controller.WaitForStatus(req, desired, closer)
	Expect(err).To(BeNil())
}

// TestController_WaitForStatus_Timeout tests Controller.WaitForStatus to end
// waiting when the given WaitTimeout expired.
func TestController_WaitForStatus_Timeout(t *testing.T) {
	RegisterTestingT(t)

	// Mocks
	newFleetMockConfig := defaultFleetMockConfig()
	newFleetMockConfig.UseTestifyMock = false
	newFleetMockConfig.UseCustomMock = true
	newFleetMockConfig.FirstCustomMockStatus = []fleet.UnitStatus{
		{
			Current: "loaded",
			Desired: "loaded",
			Machine: []fleet.MachineStatus{
				{
					ID:            "test-id",
					IP:            net.ParseIP("10.0.0.101"),
					SystemdActive: "activating",
					SystemdSub:    "start-pre",
					UnitHash:      "test-hash",
				},
			},
			Name: "test-main@1.service",
		},
	}
	newFleetMockConfig.LastCustomMockStatus = []fleet.UnitStatus{
		{
			Current: "loaded",
			Desired: "loaded",
			Machine: []fleet.MachineStatus{
				{
					ID:            "test-id",
					IP:            net.ParseIP("10.0.0.101"),
					SystemdActive: "active",
					SystemdSub:    "running",
					UnitHash:      "test-hash",
				},
			},
			Name: "test-main@1.service",
		},
	}

	c, _ := givenControllerWithConfig(newFleetMockConfig)
	c.(*controller).WaitTimeout = 0

	// Execute test
	req := api.Request{
		Group:    "test",
		SliceIDs: []string{"1"},
		Units: []api.Unit{
			{
				Name:    "test-main@1.service",
				Content: "content",
			},
		},
	}
	desired := StatusRunning
	closer := make(chan struct{}, 1)

	err := c.WaitForStatus(req, desired, closer)
	Expect(IsWaitTimeoutReached(err)).To(BeTrue()) // Because WaitForStatus is 0 nothing should happen but directly return the error
}
