package fleet

import (
	"testing"

	. "github.com/onsi/gomega"
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
