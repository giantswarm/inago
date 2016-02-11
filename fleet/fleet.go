// Package fleet implements a fleet client providing basic operations against a
// fleet endpoint through fleet's HTTP API. Higher level scheduling and
// management should be built on top of that.
package fleet

import (
	"net"
	"net/http"

	"github.com/coreos/fleet/client"
	"github.com/coreos/fleet/etcd"
	"github.com/coreos/fleet/machine"
	"github.com/coreos/fleet/registry"
	"github.com/coreos/fleet/schema"
	"github.com/coreos/fleet/unit"
)

// TODO
type Config struct{}

// TODO
func DefaultConfig() Config {
	newConfig := Config{}

	return newConfig
}

type Status struct {
	// SystemdStatus represents the status within systemd.
	SystemdStatus containerstates.State

	// FleetStatus represents the status within fleet.
	FleetStatus containerstates.State

	// HostIP represents the hosts IP where the related unit is running on. If
	// the unit is not scheduled, HostIP is empty.
	HostIP net.IP
}

// Fleet defines the interface a fleet client needs to implement to provide
// basic operations against a fleet endpoint.
type Fleet interface {
	// Create schedules a unit on the configured fleet cluster. In case the given
	// unit already exists, UnitAlreadyExistsError is returned.
	Create(name, content string) error

	// Start starts a unit on the configured fleet cluster. In case the given
	// unit is already started, UnitRunningError is returned. If there cannot be
	// any unit found, UnitNotFoundError is returned.
	Start(name string) error

	// Stop stops a unit on the configured fleet cluster. In case the given unit
	// is already stopped, UnitHaltError is returned. If there cannot be any unit
	// found, UnitNotFoundError is returned.
	Stop(name string) error

	// Delete delets a unit on the configured fleet cluster. In case the given unit
	// is not stopped, UnitRunningError is returned. If there cannot be any unit
	// found, UnitNotFoundError is returned.
	Delete(name string) error

	// GetStatus fetches the current status of a unit. If there cannot be any unit
	// found, UnitNotFoundError is returned.
	GetStatus(name string) (Status, error)
}

func NewFleet(config Config) (Fleet, error) {
	client, err := client.NewHTTPClient(config.Client, *config.Endpoint)
	if err != nil {
		return nil, mask(err)
	}

	// TODO
	newFleet := fleet{
		Config: config,
		Client: client,
	}

	return newFleet, nil
}

type fleet struct {
	Config Config
	Client *http.Client
}

// TODO
