package fleet

import (
	"net"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/coreos/fleet/machine"
	"github.com/coreos/fleet/schema"
	"github.com/stretchr/testify/mock"
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

func TestFleetSubmit_Success(t *testing.T) {
	RegisterTestingT(t)

	fleetClientMock, fleet := GivenMockedFleet()
	fleetClientMock.On("CreateUnit", mock.AnythingOfType("*schema.Unit")).Once().Return(nil, nil)
	err := fleet.Submit("unit.service", "[Unit]\n"+
		"Description=This is a test unit\n"+
		"[Service]\n"+
		"ExecStart=/bin/echo Hello World!\n")

	Expect(err).To(Not(HaveOccurred()))

	fleetClientMock.AssertCalled(
		t,
		"CreateUnit",
		mock.MatchedBy(func(unit *schema.Unit) bool {
			return unit.Name == "unit.service" &&
				unit.DesiredState == unitStateLoaded
		}),
	)
}

func TestFleetStart_Success(t *testing.T) {
	RegisterTestingT(t)

	mock, fleet := GivenMockedFleet()
	mock.On("SetUnitTargetState", "unit.service", unitStateLaunched).Once().Return(nil)

	err := fleet.Start("unit.service")

	Expect(err).To(Not(HaveOccurred()))
	mock.AssertExpectations(t)
}

func TestFleetStop_Success(t *testing.T) {
	RegisterTestingT(t)

	mock, fleet := GivenMockedFleet()
	mock.On("SetUnitTargetState", "unit.service", unitStateLoaded).Once().Return(nil)
	err := fleet.Stop("unit.service")

	Expect(err).To(Not(HaveOccurred()))
	mock.AssertExpectations(t)
}

func TestFleetDestroy_Success(t *testing.T) {
	RegisterTestingT(t)

	mock, fleet := GivenMockedFleet()
	mock.On("DestroyUnit", "unit.service").Once().Return(nil)
	err := fleet.Destroy("unit.service")

	Expect(err).To(Not(HaveOccurred()))
	mock.AssertExpectations(t)
}

func TestFleetGetStatusWithMatcher__Success(t *testing.T) {
	machineID := "12345"

	RegisterTestingT(t)

	// Mocking
	fleetClientMock, fleet := GivenMockedFleet()
	fleetClientMock.On("Units").Return([]*schema.Unit{
		{Name: "unit.service", CurrentState: unitStateLaunched, DesiredState: unitStateLaunched},
		{Name: "other.service", CurrentState: unitStateInactive, DesiredState: unitStateInactive},
	}, nil).Once()
	fleetClientMock.On("UnitStates").Return([]*schema.UnitState{
		{
			Name:               "unit.service",
			MachineID:          machineID,
			SystemdActiveState: "running",
		},
		// other.service is not scheduled
	}, nil).Once()

	fleetClientMock.On("Machines").Return([]machine.MachineState{
		{ID: machineID, PublicIP: "10.0.0.100"},
		{ID: "otherID", PublicIP: "10.0.0.254"},
	}, nil).Once()

	// Action
	matcher := func(s string) bool {
		return s == "unit.service"
	}
	status, err := fleet.GetStatusWithMatcher(matcher)

	// Assertion
	Expect(err).To(Not(HaveOccurred()))
	Expect(len(status)).To(Equal(1))
	Expect(status[0]).To(Equal(UnitStatus{
		Name:    "unit.service",
		Current: unitStateLaunched,
		Desired: unitStateLaunched,
		Machine: []MachineStatus{
			{
				ID:            machineID,
				IP:            net.ParseIP("10.0.0.100"),
				SystemdActive: "running",
			},
		},
	}))
}

func Test_Fleet_createOurStatusList(t *testing.T) {
	testCases := []struct {
		Error                error
		FoundFleetUnits      []*schema.Unit
		FoundFleetUnitStates []*schema.UnitState
		FleetMachines        []machine.MachineState
		UnitStatusList       []UnitStatus
	}{
		// This test ensures that creating our own status structures works as
		// expected.
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
		_, fleet := GivenMockedFleet()
		ourStatusList, err := fleet.createOurStatusList(testCase.FoundFleetUnits, testCase.FoundFleetUnitStates, testCase.FleetMachines)

		Expect(err).To(Not(HaveOccurred()))
		Expect(ourStatusList).To(Equal(testCase.UnitStatusList))
	}
}
