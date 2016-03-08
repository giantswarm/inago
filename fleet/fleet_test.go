package fleet

import (
	"net"
	"reflect"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/coreos/fleet/machine"
	"github.com/coreos/fleet/schema"
	"github.com/stretchr/testify/mock"
)

// Test_Fleet_DefaultConfig_Success verifies that the default config contains a
// basic valid fleet config and a valid fleet instance can be created.
func Test_Fleet_DefaultConfig_Success(t *testing.T) {
	RegisterTestingT(t)

	cfg := DefaultConfig()
	Expect(cfg.Endpoint).To(Not(BeZero()))
	Expect(cfg.Client).To(Not(BeZero()))

	newFleet, err := NewFleet(cfg)
	Expect(newFleet).To(Not(BeZero()))
	Expect(err).To(BeNil())
}

// Test_Fleet_DefaultConfig_Failure_001 verifies that a proper error will be
// thrown when the given config is invalid.
func Test_Fleet_DefaultConfig_Failure_001(t *testing.T) {
	RegisterTestingT(t)

	cfg := DefaultConfig()
	Expect(cfg.Endpoint).To(Not(BeZero()))
	Expect(cfg.Client).To(Not(BeZero()))

	cfg.Endpoint.Host = "foo"
	cfg.Endpoint.Scheme = "file"

	newFleet, err := NewFleet(cfg)
	Expect(newFleet).To(BeZero())
	Expect(IsInvalidEndpoint(err)).To(BeTrue())
}

// Test_Fleet_DefaultConfig_Failure_002 verifies that a proper error will be
// thrown when the given config is invalid.
func Test_Fleet_DefaultConfig_Failure_002(t *testing.T) {
	RegisterTestingT(t)

	cfg := DefaultConfig()
	Expect(cfg.Endpoint).To(Not(BeZero()))
	Expect(cfg.Client).To(Not(BeZero()))

	cfg.Endpoint.Scheme = "foo"

	newFleet, err := NewFleet(cfg)
	Expect(newFleet).To(BeZero())
	Expect(IsInvalidEndpoint(err)).To(BeTrue())
}

// Test_Fleet_DefaultConfig_Failure_003 verifies that the new config
// sets a new http client and does not overwrite the old one.
func Test_Fleet_DefaultConfig_Failure_003(t *testing.T) {
	RegisterTestingT(t)

	oldCfg := DefaultConfig()
	Expect(oldCfg.Endpoint).To(Not(BeZero()))
	Expect(oldCfg.Client).To(Not(BeZero()))

	newCfg := DefaultConfig()
	Expect(newCfg.Endpoint).To(Not(BeZero()))
	Expect(newCfg.Client).To(Not(BeZero()))

	Expect(oldCfg.Client).ToNot(BeIdenticalTo(newCfg.Client))
}

func givenMockedFleet() (*fleetClientMock, *fleet) {
	mock := &fleetClientMock{}
	return mock, &fleet{
		Client: mock,
		Config: DefaultConfig(),
	}
}

func givenMockedFleetWithMachines(machines []machine.MachineState) (*fleetClientMock, *fleet) {
	fleetClientMock, fleet := givenMockedFleet()
	fleetClientMock.On("Machines").Return(machines, nil)
	return fleetClientMock, fleet
}

func TestFleetSubmit_Success(t *testing.T) {
	RegisterTestingT(t)

	fleetClientMock, fleet := givenMockedFleet()
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

	mock, fleet := givenMockedFleet()
	mock.On("SetUnitTargetState", "unit.service", unitStateLaunched).Once().Return(nil)

	err := fleet.Start("unit.service")

	Expect(err).To(Not(HaveOccurred()))
	mock.AssertExpectations(t)
}

func TestFleetStop_Success(t *testing.T) {
	RegisterTestingT(t)

	mock, fleet := givenMockedFleet()
	mock.On("SetUnitTargetState", "unit.service", unitStateLoaded).Once().Return(nil)
	err := fleet.Stop("unit.service")

	Expect(err).To(Not(HaveOccurred()))
	mock.AssertExpectations(t)
}

func TestFleetDestroy_Success(t *testing.T) {
	RegisterTestingT(t)

	mock, fleet := givenMockedFleet()
	mock.On("DestroyUnit", "unit.service").Once().Return(nil)
	err := fleet.Destroy("unit.service")

	Expect(err).To(Not(HaveOccurred()))
	mock.AssertExpectations(t)
}

func TestFleetGetStatusWithMatcher__Success(t *testing.T) {
	machineID := "12345"
	machineIP := "10.0.0.100"

	RegisterTestingT(t)

	// Mocking
	fleetClientMock, fleet := givenMockedFleet()
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
		{ID: machineID, PublicIP: machineIP},
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
				IP:            net.ParseIP(machineIP),
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
					Hash:               "1234",
				},
				{
					MachineID:          "machine-ID-2",
					Name:               "name-2",
					SystemdActiveState: "systemd-active-state-2",
					Hash:               "7890",
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
							UnitHash:      "1234",
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
							UnitHash:      "7890",
						},
					},
					Name: "name-2",
				},
			},
		},
	}

	for _, testCase := range testCases {
		_, fleet := givenMockedFleet()
		ourStatusList, err := fleet.createOurStatusList(testCase.FoundFleetUnits, testCase.FoundFleetUnitStates, testCase.FleetMachines)
		if err != nil {
			t.Fatalf("Fleet.createOurStatusList returned error: %#v", err)
		}

		if !reflect.DeepEqual(ourStatusList, testCase.UnitStatusList) {
			t.Fatalf("generated status list '%#v' is not equal to expected status list '%#v'", ourStatusList, testCase.UnitStatusList)
		}
	}
}
