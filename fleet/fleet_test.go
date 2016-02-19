package fleet

import (
	"net"
	"reflect"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/coreos/fleet/machine"
	"github.com/coreos/fleet/schema"
)

// TestDefaultConfig verifies that the default config contains a basic valid fleet config
func TestDefaultConfig(t *testing.T) {
	RegisterTestingT(t)

	cfg := DefaultConfig()

	Expect(cfg.Endpoint).To(Not(BeZero()))
	Expect(cfg.Client).To(Not(BeZero()))
}

func GivenMockedFleet() (*fleetClientMock, *fleet) {
	mock := &fleetClientMock{}
	return mock, &fleet{
		Client: mock,
		Config: DefaultConfig(),
	}
}

func GivenMockedFleetWithMachines(machines []machine.MachineState) (*fleetClientMock, *fleet) {
	mock := &fleetClientMock{
		machines: machines,
	}
	return mock, &fleet{
		Client: mock,
		Config: DefaultConfig(),
	}
}

func TestFleetSubmit_Success(t *testing.T) {
	RegisterTestingT(t)

	mock, fleet := GivenMockedFleet()
	err := fleet.Submit("unit.service", "[Unit]\n"+
		"Description=This is a test unit\n"+
		"[Service]\n"+
		"ExecStart=/bin/echo Hello World!\n")

	Expect(err).To(Not(HaveOccurred()))
	Expect(len(mock.Calls())).To(Equal(1))
	call := mock.Calls()[0]
	Expect(call.Name).To(Equal("CreateUnit"))
	Expect(len(call.Args)).To(Equal(1))

	unit := call.Args[0].(*schema.Unit)
	Expect(unit.Name).To(Equal("unit.service"))
	Expect(unit.Options).To(Not(BeZero()))
	Expect(unit.DesiredState).To(Equal(unitStateLoaded))
}

func TestFleetStart_Success(t *testing.T) {
	RegisterTestingT(t)

	mock, fleet := GivenMockedFleet()
	err := fleet.Start("unit.service")

	Expect(err).To(Not(HaveOccurred()))
	Expect(mock).To(containCall("SetUnitTargetState", "unit.service", unitStateLaunched))
}

func TestFleetStop_Success(t *testing.T) {
	RegisterTestingT(t)

	mock, fleet := GivenMockedFleet()
	err := fleet.Stop("unit.service")

	Expect(err).To(Not(HaveOccurred()))
	Expect(mock).To(containCall("SetUnitTargetState", "unit.service", unitStateLoaded))
}

func TestFleetDestroy_Success(t *testing.T) {
	RegisterTestingT(t)

	mock, fleet := GivenMockedFleet()
	err := fleet.Destroy("unit.service")

	Expect(err).To(Not(HaveOccurred()))
	Expect(mock).To(containCall("DestroyUnit", "unit.service"))
}

func Test_Fleet_createOurStatusList(t *testing.T) {
	testCases := []struct {
		Error                error
		FoundFleetUnits      []*schema.Unit
		FoundFleetUnitStates []*schema.UnitState
		FleetMachines        []machine.MachineState
		UnitStatusList       []UnitStatus
	}{
		// ..
		{
			Error: nil,
			FoundFleetUnits: []*schema.Unit{
				{
					CurrentState: "current-state-1",
					DesiredState: "desired-state-1",
					MachineID:    "machine-ID-1",
					Name:         "name-1",
				},
				{
					CurrentState: "current-state-2",
					DesiredState: "desired-state-2",
					MachineID:    "machine-ID-2",
					Name:         "name-2",
				},
			},
			FoundFleetUnitStates: []*schema.UnitState{
				{
					MachineID:          "machine-ID-1",
					Name:               "name-1",
					SystemdActiveState: "systemd-active-state-1",
				},
				{
					MachineID:          "machine-ID-2",
					Name:               "name-2",
					SystemdActiveState: "systemd-active-state-2",
				},
			},
			FleetMachines: []machine.MachineState{
				{
					ID:       "machine-ID-1",
					PublicIP: "10.0.0.1",
				},
				{
					ID:       "machine-ID-2",
					PublicIP: "10.0.0.2",
				},
			},
			UnitStatusList: []UnitStatus{
				{
					Current: "current-state-1",
					Desired: "desired-state-1",
					Machine: []MachineStatus{
						{
							ID:            "machine-ID-1",
							IP:            net.ParseIP("10.0.0.1"),
							SystemdActive: "systemd-active-state-1",
						},
					},
					Name: "name-1",
				},
				{
					Current: "current-state-2",
					Desired: "desired-state-2",
					Machine: []MachineStatus{
						{
							ID:            "machine-ID-2",
							IP:            net.ParseIP("10.0.0.2"),
							SystemdActive: "systemd-active-state-2",
						},
					},
					Name: "name-2",
				},
			},
		},
	}

	for _, testCase := range testCases {
		_, fleet := GivenMockedFleetWithMachines(testCase.FleetMachines)
		ourStatusList, err := fleet.createOurStatusList(testCase.FoundFleetUnits, testCase.FoundFleetUnitStates)
		if err != nil {
			t.Fatalf("Fleet.createOurStatusList returned error: %#v", err)
		}

		if !reflect.DeepEqual(ourStatusList, testCase.UnitStatusList) {
			t.Fatalf("generated status list '%#v' is not equal to expected status list '%#v'", ourStatusList, testCase.UnitStatusList)
		}
	}
}
