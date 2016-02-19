package fleet

import (
	"testing"

	. "github.com/onsi/gomega"

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
	fleetClientMock.mock.On("CreateUnit", mock.AnythingOfType("*schema.Unit")).Once().Return(nil, nil)
	err := fleet.Submit("unit.service", "[Unit]\n"+
		"Description=This is a test unit\n"+
		"[Service]\n"+
		"ExecStart=/bin/echo Hello World!\n")

	Expect(err).To(Not(HaveOccurred()))

	fleetClientMock.mock.AssertCalled(
		t,
		"CreateUnit",
		mock.MatchedBy(func (unit *schema.Unit) bool {
			return unit.Name == "unit.service" &&
				unit.DesiredState == unitStateLoaded
		}),
	)
}

func TestFleetStart_Success(t *testing.T) {
	RegisterTestingT(t)

	mock, fleet := GivenMockedFleet()
	mock.mock.On("SetUnitTargetState", "unit.service", unitStateLaunched).Once().Return(nil)

	err := fleet.Start("unit.service")

	Expect(err).To(Not(HaveOccurred()))
	mock.mock.AssertExpectations(t)
}

func TestFleetStop_Success(t *testing.T) {
	RegisterTestingT(t)

	mock, fleet := GivenMockedFleet()
	mock.mock.On("SetUnitTargetState", "unit.service", unitStateLoaded).Once().Return(nil)
	err := fleet.Stop("unit.service")

	Expect(err).To(Not(HaveOccurred()))
	mock.mock.AssertExpectations(t)
}

func TestFleetDestroy_Success(t *testing.T) {
	RegisterTestingT(t)

	mock, fleet := GivenMockedFleet()
	mock.mock.On("DestroyUnit", "unit.service").Once().Return(nil)
	err := fleet.Destroy("unit.service")

	Expect(err).To(Not(HaveOccurred()))
	mock.mock.AssertExpectations(t)
}
