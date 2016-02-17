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

func TestFleetSubmit(t *testing.T) {
	RegisterTestingT(t)

	mock, fleet := GivenMockedFleet()
	err := fleet.Start("unit.service")

	Expect(err).To(Not(HaveOccurred()))
	Expect(mock).To(containCall("SetUnitTargetState", "unit.service", unitStateLaunched))
}
