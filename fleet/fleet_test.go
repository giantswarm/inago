package fleet

import (
	"testing"

	. "github.com/onsi/gomega"

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
