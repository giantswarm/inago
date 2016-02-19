package fleet

import (
	"fmt"

	"github.com/coreos/fleet/machine"
	"github.com/coreos/fleet/schema"
	"github.com/onsi/gomega/types"
	"github.com/stretchr/testify/mock"
)

// callRecorder returns the recorded calls
type callRecorder interface {
	TestifyMock() mock.Mock
}

// containCall returns a new GomegaMatcher to test if a given call was made.
//
//     Expect(callRecorderObj).To(containCall("functionName", arg1, arg2))
//
func containCall(name string, args ...interface{}) types.GomegaMatcher {
	return &containsCallMatcher{mock.Call{
		Method:    name,
		Arguments: args,
	}}
}

type containsCallMatcher struct {
	mock.Call
}

func (matcher *containsCallMatcher) Match(actual interface{}) (success bool, err error) {
	client := actual.(callRecorder)
	mock := client.TestifyMock()

	for _, call := range mock.Calls {
		if call.Method != matcher.Method {
			continue
		}

		_, differences := call.Arguments.Diff(matcher.Arguments)
		if differences == 0 {
			continue
		}

		return true, nil
	}
	return false, nil
}
func (matcher *containsCallMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto contain call\n\t%#v", actual, matcher.Call)
}
func (matcher *containsCallMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nnot to contain call\n\t%#v", actual, matcher.Call)
}

type fleetClientMock struct {
	mock.Mock
}

func (fleet *fleetClientMock) TestifyMock() mock.Mock {
	return fleet.Mock
}

func (fleet *fleetClientMock) Machines() ([]machine.MachineState, error) {
	args := fleet.Called()
	return args.Get(0).([]machine.MachineState), args.Error(1)
}

func (fleet *fleetClientMock) Unit(unit string) (*schema.Unit, error) {
	args := fleet.Called(unit)
	return args.Get(0).(*schema.Unit), args.Error(1)
}
func (fleet *fleetClientMock) Units() ([]*schema.Unit, error) {
	args := fleet.Called()
	return args.Get(0).([]*schema.Unit), args.Error(1)
}
func (fleet *fleetClientMock) UnitStates() ([]*schema.UnitState, error) {
	args := fleet.Called()
	return args.Get(0).([]*schema.UnitState), args.Error(1)
}

func (fleet *fleetClientMock) SetUnitTargetState(name, target string) error {
	args := fleet.Called(name, target)
	return args.Error(0)
}
func (fleet *fleetClientMock) CreateUnit(unit *schema.Unit) error {
	args := fleet.Called(unit)
	return args.Error(0)
}
func (fleet *fleetClientMock) DestroyUnit(unit string) error {
	args := fleet.Called(unit)
	return args.Error(0)
}
