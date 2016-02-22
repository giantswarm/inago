package controller

import (
	"net"
	"reflect"
	"testing"

	"github.com/giantswarm/formica/fleet"
)

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
			Input: UnitStatusList{
				{
					Current: "current-state-1",
					Desired: "desired-state-1",
					Machine: []fleet.MachineStatus{
						{
							ID:            "machine-ID-1",
							IP:            net.ParseIP("10.0.0.1"),
							SystemdActive: "systemd-active-state-1",
						},
					},
					Name: "name-1@1.service",
				},
				{
					Current: "current-state-1",
					Desired: "desired-state-1",
					Machine: []fleet.MachineStatus{
						{
							ID:            "machine-ID-1",
							IP:            net.ParseIP("10.0.0.1"),
							SystemdActive: "systemd-active-state-1",
						},
					},
					Name: "name-2@1.mount",
				},
				{
					Current: "current-state-2",
					Desired: "desired-state-2",
					Machine: []fleet.MachineStatus{
						{
							ID:            "machine-ID-2",
							IP:            net.ParseIP("10.0.0.2"),
							SystemdActive: "systemd-active-state-2",
						},
					},
					Name: "name-3@2.service",
				},
				{
					Current: "current-state-2",
					Desired: "desired-state-2",
					Machine: []fleet.MachineStatus{
						{
							ID:            "machine-ID-2",
							IP:            net.ParseIP("10.0.0.2"),
							SystemdActive: "systemd-active-state-2",
						},
					},
					Name: "name-4@2.mount",
				},
			},
			Expected: []fleet.UnitStatus{
				{
					Current: "current-state-1",
					Desired: "desired-state-1",
					Machine: []fleet.MachineStatus{
						{
							ID:            "machine-ID-1",
							IP:            net.ParseIP("10.0.0.1"),
							SystemdActive: "systemd-active-state-1",
						},
					},
					Name: "*",
				},
				{
					Current: "current-state-2",
					Desired: "desired-state-2",
					Machine: []fleet.MachineStatus{
						{
							ID:            "machine-ID-2",
							IP:            net.ParseIP("10.0.0.2"),
							SystemdActive: "systemd-active-state-2",
						},
					},
					Name: "*",
				},
			},
		},
	}

	for _, testCase := range testCases {
		output, err := testCase.Input.Group()
		if err != nil {
			t.Fatalf("Fleet.createOurStatusList returned error: %#v", err)
		}

		if !reflect.DeepEqual(output, testCase.Expected) {
			t.Fatalf("grouped status list '%#v' is not equal to expected status list '%#v'", output, testCase.Expected)
		}
	}
}
