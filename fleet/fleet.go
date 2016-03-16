// Package fleet implements a fleet client providing basic operations against a
// fleet endpoint through fleet's HTTP API. Higher level scheduling and
// management should be built on top of that.
package fleet

import (
	"strings"
	"net"
	"net/http"
	"net/url"


	"github.com/coreos/fleet/client"
	"github.com/coreos/fleet/machine"
	"github.com/coreos/fleet/schema"
	"github.com/coreos/fleet/unit"
	"golang.org/x/net/context"

	"github.com/giantswarm/inago/common"
	"github.com/giantswarm/inago/logging"
)

const (
	unitStateInactive = "inactive"
	unitStateLoaded   = "loaded"
	unitStateLaunched = "launched"
)

// Config provides all necessary and injectable configurations for a new
// fleet client.
type Config struct {
	Client   *http.Client
	Endpoint url.URL

	// Logger provides an initialised logger.
	Logger logging.Logger
}

// DefaultConfig provides a set of configurations with default values by best
// effort.
func DefaultConfig() Config {
	URL, err := url.Parse("file:///var/run/fleet.sock")
	if err != nil {
		panic(err)
	}

	newConfig := Config{
		Client:   &http.Client{},
		Endpoint: *URL,
		Logger:   logging.NewLogger(logging.DefaultConfig()),
	}

	return newConfig
}

// MachineStatus represents a unit's status scheduled on a certain machine.
type MachineStatus struct {
	// ID represents the machines fleet agent ID where the related unit is
	// running on.
	ID string

	// IP represents the machines IP where the related unit is running on.
	IP net.IP

	// SystemdActive represents the unit's systemd active state.
	SystemdActive string

	// SystemdSub represents the unit's systemd sub state.
	SystemdSub string

	// UnitHash represents a unique token to identify the content of the unitfile.
	UnitHash string
}

// UnitStatus represents the status of a unit.
type UnitStatus struct {
	// Current represents the current status within the fleet cluster.
	Current string

	// Desired represents the desired status within the fleet cluster.
	Desired string

	// Machine represents the status within a machine. For normal units that are
	// scheduled on only one machine there will be one MachineStatus returned.
	// For global units that are scheduled on multiple machines there will be
	// multiple MachineStatus returned. If a unit is not yet scheduled to any
	// machine, this will be empty.
	Machine []MachineStatus

	// Name represents the unit file name.
	Name string

	// Slice represents the slice ID. E.g. 1, or foo, or 5., etc..
	SliceID string
}

// Fleet defines the interface a fleet client needs to implement to provide
// basic operations against a fleet endpoint.
type Fleet interface {
	// Submit schedules a unit on the configured fleet cluster. This is done by
	// setting the unit's target state to loaded.
	Submit(ctx context.Context, name, content string) error

	// Start starts a unit on the configured fleet cluster. This is done by
	// setting the unit's target state to launched.
	Start(ctx context.Context, name string) error

	// Stop stops a unit on the configured fleet cluster. This is done by
	// setting the unit's target state to loaded.
	Stop(ctx context.Context, name string) error

	// Destroy delets a unit on the configured fleet cluster. This is done by
	// setting the unit's target state to inactive.
	Destroy(ctx context.Context, name string) error

	// GetStatus fetches the current status of a unit. If the unit cannot be
	// found, an error that you can identify using IsUnitNotFound is returned.
	GetStatus(ctx context.Context, name string) (UnitStatus, error)

	// GetStatusWithMatcher returns a []UnitStatus, with an element for
	// each unit where the given matcher returns true.
	GetStatusWithMatcher(func(string) bool) ([]UnitStatus, error)
}

// NewFleet creates a new Fleet that is configured with the given settings.
//
//   newConfig := fleet.DefaultConfig()
//   newConfig.Endpoint = myCustomEndpoint
//   newFleet := fleet.NewFleet(newConfig)
//
func NewFleet(config Config) (Fleet, error) {
	var trans http.RoundTripper

	switch config.Endpoint.Scheme {
	case "unix", "file":
		if config.Endpoint.Host != "" {
			// This commonly happens if the user misses the leading slash after the
			// scheme. For example, "unix://var/run/fleet.sock" would be parsed as
			// host "var".
			return nil, maskAnyf(invalidEndpointError, "cannot connect to host %q with scheme %q", config.Endpoint.Host, config.Endpoint.Scheme)
		}

		// The Path field is only used for dialing and should not be used when
		// building any further HTTP requests.
		sockPath := config.Endpoint.Path
		config.Endpoint.Path = ""

		// http.Client doesn't support the schemes "unix" or "file", but it
		// is safe to use "http" as dialFunc ignores it anyway.
		config.Endpoint.Scheme = "http"

		// The Host field is not used for dialing, but will be exposed in debug logs.
		config.Endpoint.Host = "domain-sock"

		trans = &http.Transport{
			Dial: func(s, t string) (net.Conn, error) {
				// http.Client does not natively support dialing a unix domain socket,
				// so the dial function must be overridden.
				return net.Dial("unix", sockPath)
			},
		}
	case "http", "https":
		trans = http.DefaultTransport
	default:
		return nil, maskAnyf(invalidEndpointError, "invalid scheme %q", config.Endpoint.Scheme)
	}

	config.Client.Transport = trans

	client, err := client.NewHTTPClient(config.Client, config.Endpoint)
	if err != nil {
		return nil, maskAny(err)
	}

	newFleet := fleet{
		Config: config,
		Client: client,
	}

	return newFleet, nil
}

type fleet struct {
	Config Config
	Client client.API
}

func (f fleet) Submit(ctx context.Context, name, content string) error {
	f.Config.Logger.Debug(ctx, "fleet: submitting unit '%v'", name)

	unitFile, err := unit.NewUnitFile(content)
	if err != nil {
		return maskAny(err)
	}

	unit := &schema.Unit{
		Name:         name,
		Options:      schema.MapUnitFileToSchemaUnitOptions(unitFile),
		DesiredState: "loaded",
	}

	err = f.Client.CreateUnit(unit)
	if err != nil {
		return maskAny(err)
	}

	return nil
}

func (f fleet) Start(ctx context.Context, name string) error {
	f.Config.Logger.Debug(ctx, "fleet: starting unit '%v'", name)

	err := f.Client.SetUnitTargetState(name, unitStateLaunched)
	if err != nil {
		return maskAny(err)
	}

	return nil
}

func (f fleet) Stop(ctx context.Context, name string) error {
	f.Config.Logger.Debug(ctx, "fleet: stopping unit '%v'", name)

	err := f.Client.SetUnitTargetState(name, unitStateLoaded)
	if err != nil {
		return maskAny(err)
	}

	return nil
}

func (f fleet) Destroy(ctx context.Context, name string) error {
	f.Config.Logger.Debug(ctx, "fleet: destroying unit '%v'", name)

	err := f.Client.DestroyUnit(name)
	if err != nil {
		return maskAny(err)
	}

	return nil
}

func (f fleet) GetStatus(ctx context.Context, name string) (UnitStatus, error) {
	f.Config.Logger.Debug(ctx, "fleet: getting status of unit '%v'", name)

	matcher := func(s string) bool {
		return name == s
	}
	unitStatus, err := f.GetStatusWithMatcher(matcher)
	if err != nil {
		return UnitStatus{}, maskAny(err)
	}

	if len(unitStatus) != 1 {
		return UnitStatus{}, maskAny(invalidUnitStatusError)
	}

	return unitStatus[0], nil
}

// GetStatusWithMatcher returns a []UnitStatus, with an element for
// each unit where the given matcher returns true.
func (f fleet) GetStatusWithMatcher(matcher func(s string) bool) ([]UnitStatus, error) {
	// Lookup fleet cluster state.
	fleetUnits, err := f.Client.Units()
	if err != nil {
		return []UnitStatus{}, maskAny(err)
	}
	foundFleetUnits := []*schema.Unit{}
	for _, fu := range fleetUnits {
		if matcher(fu.Name) {
			foundFleetUnits = append(foundFleetUnits, fu)
		}
	}

	// Return not found error if there is no unit as requested.
	if len(foundFleetUnits) == 0 {
		return []UnitStatus{}, maskAny(unitNotFoundError)
	}

	// Lookup machine states.
	fleetUnitStates, err := f.Client.UnitStates()
	if err != nil {
		return []UnitStatus{}, maskAny(err)
	}
	var foundFleetUnitStates []*schema.UnitState
	for _, fus := range fleetUnitStates {
		if matcher(fus.Name) {
			foundFleetUnitStates = append(foundFleetUnitStates, fus)
		}
	}

	// Lookup machines
	machineStates, err := f.Client.Machines()
	if err != nil {
		return nil, maskAny(err)
	}

	// Create our own unit status.
	ourStatusList, err := mapFleetStateToUnitStatusList(foundFleetUnits, foundFleetUnitStates, machineStates)
	if err != nil {
		return []UnitStatus{}, maskAny(err)
	}

	return ourStatusList, nil
}

func ipFromUnitState(unitState *schema.UnitState, machineStates []machine.MachineState) (net.IP, error) {
	for _, ms := range machineStates {
		if unitState.MachineID == ms.ID {
			return net.ParseIP(ms.PublicIP), nil
		}
	}

	return nil, maskAny(ipNotFoundError)
}

func mapFleetStateToUnitStatusList(foundFleetUnits []*schema.Unit, foundFleetUnitStates []*schema.UnitState, machines []machine.MachineState) ([]UnitStatus, error) {
	ourStatusList := []UnitStatus{}

	for _, ffu := range foundFleetUnits {
		ID, err := common.SliceID(ffu.Name)
		if err != nil {
			return nil, maskAny(invalidUnitStatusError)
		}

		ourUnitStatus := UnitStatus{
			Current: ffu.CurrentState,
			Desired: ffu.DesiredState,
			Machine: []MachineStatus{},
			Name:    ffu.Name,
			SliceID: ID,
		}

		// FLEET-WEIRDNESS: In case of global units, the CurrentState seems to be always "inactive"
		// To make the output a bit nicer, we overwrite it with DesiredState
		if isFleetGlobalUnit(ffu.Options) {
			ourUnitStatus.Current = ourUnitStatus.Desired
		}

		for _, ffus := range foundFleetUnitStates {
			if ffu.Name != ffus.Name {
				continue
			}

			IP, err := ipFromUnitState(ffus, machines)
			if err != nil {
				return []UnitStatus{}, maskAny(err)
			}
			ourMachineStatus := MachineStatus{
				ID:            ffus.MachineID,
				IP:            IP,
				SystemdActive: ffus.SystemdActiveState,
				SystemdSub:    ffus.SystemdSubState,
				UnitHash:      ffus.Hash,
			}
			ourUnitStatus.Machine = append(ourUnitStatus.Machine, ourMachineStatus)
		}
		ourStatusList = append(ourStatusList, ourUnitStatus)
	}

	return ourStatusList, nil
}

func isFleetGlobalUnit(options []*schema.UnitOption) bool {
	for _, option := range options {
		if strings.EqualFold(option.Section, "X-Fleet") &&
		 strings.EqualFold(option.Name, "Global") &&
		 strings.EqualFold(option.Value, "true") {
			return true
		}
	}
	return false
}
