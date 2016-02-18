package fleet

import (
	"fmt"
	"reflect"

	"github.com/coreos/fleet/machine"
	"github.com/coreos/fleet/schema"
	"github.com/onsi/gomega/types"
)

// callRecorder returns the recorded calls
type callRecorder interface {
	Calls() []mockCall
}

type mockCall struct {
	Name string
	Args []interface{}
}

// containCall returns a new GomegaMatcher to test if a given call was made.
//
//     Expect(callRecorderObj).To(containCall("functionName", arg1, arg2))
//
func containCall(name string, args ...interface{}) types.GomegaMatcher {
	return &containsCallMatcher{mockCall{
		name, args,
	}}
}

type containsCallMatcher struct {
	expectedCall mockCall
}

func (matcher *containsCallMatcher) Match(actual interface{}) (success bool, err error) {
	client := actual.(callRecorder)

	for _, call := range client.Calls() {
		if reflect.DeepEqual(matcher.expectedCall, call) {
			return true, nil
		}
	}
	return false, nil
}
func (matcher *containsCallMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto contain call\n\t%#v", actual, matcher.expectedCall)
}
func (matcher *containsCallMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nnot to contain call\n\t%#v", actual, matcher.expectedCall)
}

type fleetClientMock struct {
	calls []mockCall
}

func (fleet *fleetClientMock) Calls() []mockCall {
	return fleet.calls
}

func (fleet *fleetClientMock) Machines() ([]machine.MachineState, error) {
	fleet.calls = append(fleet.calls, mockCall{"Machines", []interface{}{}})
	return nil, nil
}

func (fleet *fleetClientMock) Unit(unit string) (*schema.Unit, error) {
	fleet.calls = append(fleet.calls, mockCall{"Unit", []interface{}{unit}})
	return nil, nil
}
func (fleet *fleetClientMock) Units() ([]*schema.Unit, error) {
	fleet.calls = append(fleet.calls, mockCall{"Units", []interface{}{}})
	return nil, nil
}
func (fleet *fleetClientMock) UnitStates() ([]*schema.UnitState, error) {
	fleet.calls = append(fleet.calls, mockCall{"UnitStates", []interface{}{}})
	return nil, nil
}

func (fleet *fleetClientMock) SetUnitTargetState(name, target string) error {
	fleet.calls = append(fleet.calls, mockCall{"SetUnitTargetState", []interface{}{name, target}})
	return nil
}
func (fleet *fleetClientMock) CreateUnit(unit *schema.Unit) error {
	fleet.calls = append(fleet.calls, mockCall{"CreateUnit", []interface{}{unit}})
	return nil
}
func (fleet *fleetClientMock) DestroyUnit(unit string) error {
	fleet.calls = append(fleet.calls, mockCall{"DestroyUnit", []interface{}{unit}})
	return nil
}
