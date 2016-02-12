// Package fleet implements a fleet client providing basic operations against a
// fleet endpoint through fleet's HTTP API. Higher level scheduling and
// management should be built on top of that.
package fleet

import (
	"net"
	"net/http"
	"net/url"

	"github.com/coreos/fleet/client"
	"github.com/coreos/fleet/etcd"
	"github.com/coreos/fleet/machine"
	"github.com/coreos/fleet/registry"
	"github.com/coreos/fleet/schema"
	"github.com/coreos/fleet/unit"
)

type Config struct {
	Client   *http.Client
	Endpoint url.URL
}

func DefaultConfig() Config {
	URL, err := url.Parse("file:///var/run/fleet.sock")
	if err != nil {
		panic(err)
	}

	newConfig := Config{
		Client:   http.DefaultClient,
		Endpoint: *URL,
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
}

// Fleet defines the interface a fleet client needs to implement to provide
// basic operations against a fleet endpoint.
type Fleet interface {
	// Submit schedules a unit on the configured fleet cluster. In case the given
	// unit already exists, UnitAlreadyExistsError is returned.
	Submit(name, content string) error

	// Start starts a unit on the configured fleet cluster. In case the given
	// unit is already started, UnitRunningError is returned. If there cannot be
	// any unit found, UnitNotFoundError is returned.
	Start(name string) error

	// Stop stops a unit on the configured fleet cluster. In case the given unit
	// is already stopped, UnitHaltError is returned. If there cannot be any unit
	// found, UnitNotFoundError is returned.
	Stop(name string) error

	// Destroy delets a unit on the configured fleet cluster. In case the given
	// unit is not stopped, UnitRunningError is returned. If there cannot be any
	// unit found, UnitNotFoundError is returned.
	Destroy(name string) error

	// GetStatus fetches the current status of a unit. If there cannot be any unit
	// found, UnitNotFoundError is returned.
	GetStatus(name string) (UnitStatus, error)
}

func NewFleet(config Config) (Fleet, error) {
	client, err := client.NewHTTPClient(config.Client, config.Endpoint)
	if err != nil {
		return nil, mask(err)
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

func (f fleet) Submit(name, content string) error {
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

func (f fleet) Start(name string) error {
	err := f.Client.SetUnitTargetState(name, "launched")
	if err != nil {
		return maskAny(err)
	}

	return nil
}

func (f fleet) Stop(name string) error {
	err := f.Client.SetUnitTargetState(name, "loaded")
	if err != nil {
		return maskAny(err)
	}

	return nil
}

func (f fleet) Destroy(name string) error {
	err := f.Client.SetUnitTargetState(name, "inactive")
	if err != nil {
		return maskAny(err)
	}

	return nil
}

func (f fleet) GetStatus(name string) (UnitStatus, error) {
	// Lookup fleet cluster state.
	fleetUnits := f.Client.Units()
	if err != nil {
		return UnitStatus{}, maskAny(err)
	}
	var foundFleetUnit *schema.Unit
	for _, fu := range fleetUnits {
		if name == fu.Name {
			foundFleetUnit = fu
			break
		}
	}

	// Lookup machine states.
	fleetUnitStates := f.Client.UnitStates()
	if err != nil {
		return UnitStatus{}, maskAny(err)
	}
	var foundFleetUnitStates []*schema.UnitState
	for _, fus := range fleetUnitStates {
		if name == fus.Name {
			foundFleetUnitStates = append(foundFleetUnitStates, fus)
		}
	}

	// Aggregate our own unit status.
	ourUnitStatus := UnitStatus{
		Current: foundFleetUnit.CurrentState,
		Desired: foundFleetUnit.DesiredState,
		Machine: []MachineStatus{},
	}
	for _, ffus := range foundFleetUnitStates {
		IP, err := f.ipFromUnitState(ffus)
		if err != nil {
			return UnitStatus{}, maskAny(err)
		}
		ourMachineStatus := MachineStatus{
			ID:            ffus.MachineID,
			IP:            IP,
			SystemdActive: ffus.SystemdActiveState,
		}
		ourUnitStatus.Machine = append(ourUnitStatus.Machine, ourMachineStatus)
	}

	return ourUnitStatus, nil
}

func ipFromUnitState(unitState *schema.UnitState) (net.IP, error) {
	machineStates, err := f.Client.Machines()
	if err != nil {
		return nil, maskAny(err)
	}

	var ip net.IP
	for _, ms := range machineStates {
		if unitState.MachineID == ms.ID {
			return net.ParseIP(ms.PublicIP)
		}
	}

	return nil, maskAny(ipNotFoundError)
}
