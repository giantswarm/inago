// Package controller implements a controller client providing basic operations against a
// controller endpoint through controller's HTTP API. Higher level scheduling and
// management should be built on top of that.
package controller

import (
	"github.com/giantswarm/formica/fleet"
)

type Config struct {
	Fleet fleet.Fleet
}

func DefaultConfig() Config {
	newFleetConfig := fleet.DefaultConfig()
	newFleet, err := fleet.NewFleet(newFleetConfig)
	if err != nil {
		panic(err)
	}

	newConfig := Config{
		Fleet: newFleet,
	}

	return newConfig
}

// Controller defines the interface a controller needs to implement to provide
// operations for groups of unit files against a fleet cluster.
type Controller interface {
	// Submit schedules a group on the configured fleet cluster. This is done by
	// setting the units target state to loaded.
	Submit(group string) error

	// Start starts a group on the configured fleet cluster. This is done by
	// setting the units target state to launched.
	Start(group string) error

	// Stop stops a group on the configured fleet cluster. This is done by
	// setting the unit's target state to loaded.
	Stop(group string) error

	// Destroy delets a group on the configured fleet cluster. This is done by
	// setting the unit's target state to inactive.
	Destroy(group string) error

	// GetStatus fetches the current status of a group. If the unit cannot be
	// found, an error that you can identify using IsUnitNotFound is returned.
	GetStatus(group string) (fleet.UnitStatus, error)
}

func NewController(config Config) Controller {
	newController := controller{
		Config: config,
	}

	return newController
}

type controller struct {
	Config
}

func (c controller) Submit(group string) error {
	// TODO turn group into unit files

	// TODO submit each unit file
	unit := ""
	content := ""
	err := c.Fleet.Submit(unit, content)
	if err != nil {
		return maskAny(err)
	}

	// TODO retry operations

	return nil
}

// TODO
func (c controller) Start(group string) error {
	return nil
}

// TODO
func (c controller) Stop(group string) error {
	return nil
}

// TODO
func (c controller) Destroy(group string) error {
	return nil
}

// TODO
func (c controller) GetStatus(group string) (fleet.UnitStatus, error) {
	return fleet.UnitStatus{}, nil
}
