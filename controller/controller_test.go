package controller

import (
	"testing"

	"github.com/juju/errgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	"github.com/giantswarm/formica/fleet"
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
	}

	for _, testCase := range testCases {
		output, err := testCase.Input.ExtendSlices()
		if err != nil {
			t.Fatalf("Request.ExtendSlices returned error: %#v", err)
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

// GivenController returns a controller where the fleet backend is replaced
// with a mock.
func GivenController() (Controller, *fleetMock) {
	fleetMock := fleetMock{}
	cfg := Config{
		Fleet: &fleetMock,
	}
	return NewController(cfg), &fleetMock
}

func TestController_Start(t *testing.T) {
	RegisterTestingT(t)

	// Mocks
	controller, fleetMock := GivenController()
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
	err := controller.Start(req)

	// Assert
	Expect(err).To(BeNil())
	mock.AssertExpectationsForObjects(t, fleetMock.Mock)
}

func TestController_Status_ErrorOnMismatchingSliceIDs(t *testing.T) {
	RegisterTestingT(t)

	// Mocks
	controller, fleetMock := GivenController()
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
		Group: "test",
		SliceIDs: []string{"1","2"},
	})

	// Assert
	Expect(IsUnitSliceNotFound(err)).To(Equal(true))
	Expect(status).To(BeEmpty())
	mock.AssertExpectationsForObjects(t, fleetMock.Mock)
}
