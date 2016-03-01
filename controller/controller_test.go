package controller

import (
	"testing"

	"github.com/juju/errgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	"github.com/giantswarm/formica/fleet"
	"github.com/giantswarm/formica/task"
)

func Test_Request_ExtendSlices(t *testing.T) {
	testCases := []struct {
		Input    Request
		Expected Request
	}{
		// This test ensures that the request is not manipulated when no slice IDs
		// are given.
		{
			Input: Request{
				SliceIDs: []string{},
				Units: []Unit{
					{
						Name:    "unit@.service",
						Content: "some unit content",
					},
				},
			},
			Expected: Request{
				SliceIDs: []string{},
				Units: []Unit{
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
			Input: Request{
				SliceIDs: []string{"1"},
				Units: []Unit{
					{
						Name:    "unit@.service",
						Content: "some unit content",
					},
				},
			},
			Expected: Request{
				SliceIDs: []string{"1"},
				Units: []Unit{
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
			Input: Request{
				SliceIDs: []string{"1", "2"},
				Units: []Unit{
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
			Expected: Request{
				SliceIDs: []string{"1", "2"},
				Units: []Unit{
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
			Input: Request{
				SliceIDs: []string{"3", "5", "foo"},
				Units: []Unit{
					{
						Name:    "unit@.service",
						Content: "some unit content",
					},
				},
			},
			Expected: Request{
				SliceIDs: []string{"3", "5", "foo"},
				Units: []Unit{
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
			Input: Request{
				SliceIDs: nil,
				Units: []Unit{
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
			Expected: Request{
				SliceIDs: nil,
				Units: []Unit{
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
		Request        Request
		UnitStatusList []fleet.UnitStatus
	}{
		// This test ensures that validating a unit status list against given slice
		// IDs works as expected.
		{
			Error: nil,
			Request: Request{
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
			Request: Request{
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
		InputRequest  Request
		Output        bool
	}{
		{
			InputUnitName: "demo-main@1.service",
			InputRequest: Request{
				Group:    "demo",
				SliceIDs: []string{"1", "2"},
			},
			Output: true,
		},
		{
			InputUnitName: "demo-main@1.service",
			InputRequest: Request{
				Group:    "demo",
				SliceIDs: []string{"3"},
			},
			Output: false,
		},
		{
			InputUnitName: "other-main@1.service",
			InputRequest: Request{
				Group:    "demo",
				SliceIDs: []string{"1"},
			},
			Output: false,
		},
		{
			InputUnitName: "other-main@1.service",
			InputRequest: Request{
				Group:    "demo",
				SliceIDs: []string{"2"},
			},
			Output: false,
		},

		{
			InputUnitName: "demo-main.service",
			InputRequest: Request{
				Group:    "demo",
				SliceIDs: nil,
			},
			Output: true,
		},

		{
			InputUnitName: "other-main.service",
			InputRequest: Request{
				Group:    "demo",
				SliceIDs: []string{"1", "2"},
			},
			Output: false,
		},
		{
			InputUnitName: "other-main.service",
			InputRequest: Request{
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
func givenController() (Controller, *fleetMock) {
	fleetMock := fleetMock{}

	newTaskServiceConfig := task.DefaultConfig()
	newTaskService := task.NewTaskService(newTaskServiceConfig)

	cfg := Config{
		Fleet:       &fleetMock,
		TaskService: newTaskService,
	}
	return NewController(cfg), &fleetMock
}

func TestController_Submit_Error(t *testing.T) {
	RegisterTestingT(t)

	controller, fleetMock := givenController()

	// Execute
	req := Request{
		Group:    "single",
		SliceIDs: nil,
		Units:    []Unit{}, // Intentionally left empty!
	}

	closer := make(chan struct{}, 1)
	closer <- struct{}{}
	task, err := controller.Submit(req, closer)

	// Assert
	Expect(task).To(BeNil())
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(Equal("invalid argument: Units must not be empty"))
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
	).Once()
	fleetMock.On("Start", "test-main@1.service").Return(nil).Once()
	fleetMock.On("Start", "test-sidekick@1.service").Return(nil).Once()

	// Execute test
	req := Request{
		Group:    "test",
		SliceIDs: []string{"1"},
	}
	closer := make(chan struct{}, 1)
	closer <- struct{}{}
	taskObject, err := controller.Start(req, closer)
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
	).Once()
	fleetMock.On("Destroy", "test-main@1.service").Return(nil).Once()
	fleetMock.On("Destroy", "test-sidekick@1.service").Return(nil).Once()

	// Execute test
	req := Request{
		Group:    "test",
		SliceIDs: []string{"1"},
	}
	closer := make(chan struct{}, 1)
	closer <- struct{}{}
	taskObject, err := controller.Destroy(req, closer)
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
	).Once()
	fleetMock.On("Stop", "test-main@1.service").Return(nil).Once()
	fleetMock.On("Stop", "test-sidekick@1.service").Return(nil).Once()

	// Execute test
	req := Request{
		Group:    "test",
		SliceIDs: []string{"1"},
	}
	closer := make(chan struct{}, 1)
	closer <- struct{}{}
	taskObject, err := controller.Stop(req, closer)
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
	status, err := controller.GetStatus(Request{
		Group:    "test",
		SliceIDs: []string{"1", "2"},
	})

	// Assert
	Expect(IsUnitSliceNotFound(err)).To(Equal(true))
	Expect(status).To(BeEmpty())
	mock.AssertExpectationsForObjects(t, fleetMock.Mock)
}
