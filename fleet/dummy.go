package fleet

import (
	"sync"

	"golang.org/x/net/context"

	"github.com/giantswarm/inago/logging"
)

// DummyConfig holds configuration for the DummyFleet struct.
type DummyConfig struct {
	Logger logging.Logger
}

// DummyFleet is an implementation of the Fleet interface,
// that is primarily intended to be used for testing.
type DummyFleet struct {
	Config DummyConfig
	Units  map[string]UnitStatus
	Mutex  sync.Mutex
}

// DefaultDummyConfig returns a best-effort configuration for the DummyFleet struct.
func DefaultDummyConfig() DummyConfig {
	return DummyConfig{
		Logger: logging.NewLogger(logging.DefaultConfig()),
	}
}

// NewDummyFleet returns a DummyFleet, given a DummyConfig.
func NewDummyFleet(DummyConfig) *DummyFleet {
	return &DummyFleet{
		Config: DefaultDummyConfig(),
		Units:  make(map[string]UnitStatus),
	}
}

// Submit creates a UnitStatus, and stores it.
func (f *DummyFleet) Submit(ctx context.Context, name, content string) error {
	f.Config.Logger.Debug(ctx, "dummy fleet: submit %v %v", name, content)

	f.Mutex.Lock()
	defer f.Mutex.Unlock()

	f.Units[name] = UnitStatus{
		Current: unitStateLoaded,
		Desired: unitStateLoaded,
		Name:    name,
	}

	return nil
}

// Start sets the Current and Desired state of the stored UnitStatus
// to unitStateLaunched.
func (f *DummyFleet) Start(ctx context.Context, name string) error {
	f.Config.Logger.Debug(ctx, "dummy fleet: start %v", name)

	f.Mutex.Lock()
	defer f.Mutex.Unlock()

	if _, ok := f.Units[name]; !ok {
		return maskAny(unitNotFoundError)
	}

	unitStatus, _ := f.Units[name]

	unitStatus.Current = unitStateLaunched
	unitStatus.Desired = unitStateLaunched

	f.Units[name] = unitStatus

	return nil
}

// Stop sets the Current and Desired state of the stored UnitStatus
// to unitStateLoaded.
func (f *DummyFleet) Stop(ctx context.Context, name string) error {
	f.Config.Logger.Debug(ctx, "dummy fleet: stop %v", name)

	f.Mutex.Lock()
	defer f.Mutex.Unlock()

	if _, ok := f.Units[name]; !ok {
		return maskAny(unitNotFoundError)
	}

	unitStatus, _ := f.Units[name]

	unitStatus.Current = unitStateLoaded
	unitStatus.Desired = unitStateLoaded

	f.Units[name] = unitStatus

	return nil
}

// Destroy remove the UnitStatus from the internal store.
func (f *DummyFleet) Destroy(ctx context.Context, name string) error {
	f.Config.Logger.Debug(ctx, "dummy fleet: destroy %v", name)

	f.Mutex.Lock()
	defer f.Mutex.Unlock()

	if _, ok := f.Units[name]; !ok {
		return maskAny(unitNotFoundError)
	}

	delete(f.Units, name)

	return nil
}

// GetStatus returns the UnitStatus for the given name.
func (f *DummyFleet) GetStatus(ctx context.Context, name string) (UnitStatus, error) {
	f.Config.Logger.Debug(ctx, "dummy fleet: get status %v", name)

	f.Mutex.Lock()
	defer f.Mutex.Unlock()

	unitStatus, ok := f.Units[name]
	if !ok {
		return UnitStatus{}, maskAny(unitNotFoundError)
	}

	return unitStatus, nil
}

// GetStatusWithMatcher returns all UnitStatus that match.
func (f *DummyFleet) GetStatusWithMatcher(m func(string) bool) ([]UnitStatus, error) {
	f.Config.Logger.Debug(context.Background(), "dummy fleet: get status with matcher")

	f.Mutex.Lock()
	defer f.Mutex.Unlock()

	if len(f.Units) == 0 {
		return []UnitStatus{}, maskAny(unitNotFoundError)
	}

	unitStatusList := []UnitStatus{}
	for _, unitStatus := range f.Units {
		if m(unitStatus.Name) {
			unitStatusList = append(unitStatusList, unitStatus)
		}
	}

	return unitStatusList, nil
}
