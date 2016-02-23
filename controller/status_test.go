package controller

import (
	"fmt"
	"net"
	"reflect"
	"testing"

	"github.com/giantswarm/formica/fleet"
)

func inputUnitStatusList(configs ...map[string][]string) UnitStatusList {
	unitStatusList := UnitStatusList{}

	i := 0
	for _, c := range configs {
		for j, sliceID := range c["sliceIDs"] {
			state := c["states"][j]
			i++

			unitStatus := fleet.UnitStatus{
				Current: fmt.Sprintf("current-state-%s", state),
				Desired: fmt.Sprintf("desired-state-%s", state),
				Machine: []fleet.MachineStatus{
					{
						ID:            fmt.Sprintf("machine-ID-%s", sliceID),
						IP:            net.ParseIP(fmt.Sprintf("10.0.0.%s", sliceID)),
						SystemdActive: fmt.Sprintf("systemd-active-state-%s", state),
					},
				},
				Name: fmt.Sprintf("name-%d@%s.service", i, sliceID),
			}

			unitStatusList = append(unitStatusList, unitStatus)
		}
	}

	return unitStatusList
}

func expectedUnitStatusList(configs ...map[string][]string) []fleet.UnitStatus {
	unitStatusList := []fleet.UnitStatus{}

	i := 0
	for _, c := range configs {
		for j, sliceID := range c["sliceIDs"] {
			state := c["states"][j]
			name := c["names"][j]
			i++

			unitStatus := fleet.UnitStatus{
				Current: fmt.Sprintf("current-state-%s", state),
				Desired: fmt.Sprintf("desired-state-%s", state),
				Machine: []fleet.MachineStatus{
					{
						ID:            fmt.Sprintf("machine-ID-%s", sliceID),
						IP:            net.ParseIP(fmt.Sprintf("10.0.0.%s", sliceID)),
						SystemdActive: fmt.Sprintf("systemd-active-state-%s", state),
					},
				},
				Name: name,
			}
			unitStatusList = append(unitStatusList, unitStatus)
		}
	}

	return unitStatusList
}

func Test_UnitStatusList_Group(t *testing.T) {
	testCases := []struct {
		Error    error
		Input    UnitStatusList
		Expected []fleet.UnitStatus
	}{
		// This test ensures that creating our own status structures works as
		// expected.
		{
			Error: nil,
			Input: inputUnitStatusList(
				map[string][]string{"sliceIDs": []string{"1", "1"}, "states": []string{"1", "1"}},
				map[string][]string{"sliceIDs": []string{"2", "2"}, "states": []string{"2", "2"}},
			),
			Expected: expectedUnitStatusList(
				map[string][]string{"sliceIDs": []string{"1"}, "states": []string{"1"}, "names": []string{"*"}},
				map[string][]string{"sliceIDs": []string{"2"}, "states": []string{"2"}, "names": []string{"*"}},
			),
		},

		// This test ensures that different states expand the status list.
		{
			Error: nil,
			Input: inputUnitStatusList(
				map[string][]string{"sliceIDs": []string{"1", "1"}, "states": []string{"1", "2"}}, // the last state differs
				map[string][]string{"sliceIDs": []string{"2", "2"}, "states": []string{"2", "2"}},
			),
			Expected: expectedUnitStatusList(
				map[string][]string{"sliceIDs": []string{"1", "1"}, "states": []string{"1", "2"}, "names": []string{"name-1@1.service", "name-2@1.service"}}, // the states expand
				map[string][]string{"sliceIDs": []string{"2"}, "states": []string{"2"}, "names": []string{"*"}},
			),
		},
	}

	for i, testCase := range testCases {
		output, err := testCase.Input.Group()
		if err != nil {
			t.Fatalf("UnitStatusList.Group returned error: %#v", err)
		}

		if !reflect.DeepEqual(output, testCase.Expected) {
			t.Fatalf("test case %d: grouped status list '%#v' is not equal to expected status list '%#v'", i+1, output, testCase.Expected)
		}
	}
}
