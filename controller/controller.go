// Package controller implements a controller client providing basic operations against a
// controller endpoint through controller's HTTP API. Higher level scheduling and
// management should be built on top of that.
package controller

import (
	"fmt"
	"regexp"

	"github.com/giantswarm/formica/fleet"
)

// Config provides all necessary and injectable configurations for a new
// controller.
type Config struct {
	Fleet fleet.Fleet
}

// DefaultConfig provides a set of configurations with default values by best
// effort.
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
	// setting the state of the units in the group to loaded.
	Submit(req Request) error

	// Start starts a group on the configured fleet cluster. This is done by
	// setting the state of the units in the group to launched.
	Start(req Request) error

	// Stop stops a group on the configured fleet cluster. This is done by
	// setting the state of the units in the group to loaded.
	Stop(req Request) error

	// Destroy delets a group on the configured fleet cluster. This is done by
	// setting the state of the units in the group to inactive.
	Destroy(req Request) error

	// GetStatus fetches the current status of a group. If the unit cannot be
	// found, an error that you can identify using IsUnitNotFound is returned.
	GetStatus(req Request) ([]fleet.UnitStatus, error)
}

// NewController creates a new Controller that is configured with the given
// settings.
//
//   newConfig := controller.DefaultConfig()
//   newConfig.Fleet = myCustomFleetClient
//   newController := controller.NewController(newConfig)
//
func NewController(config Config) Controller {
	newController := controller{
		Config: config,
	}

	return newController
}

type controller struct {
	Config
}

// Unit represents a systemd unit file.
type Unit struct {
	// Name is something like "appd@.service". It needs to be extended using the
	// slice ID before submitting to fleet.
	Name string

	// Content represents normal systemd unit file content.
	Content string
}

// Request represents a controller request. This is used to process some action
// on the controller.
type Request struct {
	// SliceIDs contains the IDs to create. IDs can be "1", "first", "whatever",
	// "5", etc..
	SliceIDs []string

	// Units represents a list of unit files that is supposed to be extended
	// using the provided slice IDs.
	Units []Unit
}

var unitExp = regexp.MustCompile("@.service")

// ExtendSlices extends unit files with respect to the given slice IDs. Having
// slice IDs "1" and "2" and having unit files "foo@.service" and
// "bar@.service" results in the following extended unit files.
//
// 	 foo@1.service
// 	 bar@1.service
// 	 foo@2.service
// 	 bar@2.service
//
func (r Request) ExtendSlices() (Request, error) {
	newRequest := Request{
		SliceIDs: r.SliceIDs,
		Units:    []Unit{},
	}

	for _, sliceID := range r.SliceIDs {
		for _, unit := range r.Units {
			newUnit := unit
			newUnit.Name = unitExp.ReplaceAllString(newUnit.Name, fmt.Sprintf("@%s.service", sliceID))
			newRequest.Units = append(newRequest.Units, newUnit)
		}
	}

	return newRequest, nil
}

func (c controller) Submit(req Request) error {
	extended, err := req.ExtendSlices()
	if err != nil {
		return maskAny(err)
	}

	for _, unit := range extended.Units {
		err := c.Fleet.Submit(unit.Name, unit.Content)
		if err != nil {
			return maskAny(err)
		}
	}

	// TODO retry operations

	return nil
}

func (c controller) Start(req Request) error {
	extended, err := req.ExtendSlices()
	if err != nil {
		return maskAny(err)
	}

	for _, unit := range extended.Units {
		err := c.Fleet.Start(unit.Name)
		if err != nil {
			return maskAny(err)
		}
	}

	// TODO retry operations

	return nil
}

func (c controller) Stop(req Request) error {
	extended, err := req.ExtendSlices()
	if err != nil {
		return maskAny(err)
	}

	for _, unit := range extended.Units {
		err := c.Fleet.Stop(unit.Name)
		if err != nil {
			return maskAny(err)
		}
	}

	// TODO retry operations

	return nil
}

func (c controller) Destroy(req Request) error {
	extended, err := req.ExtendSlices()
	if err != nil {
		return maskAny(err)
	}

	for _, unit := range extended.Units {
		err := c.Fleet.Destroy(unit.Name)
		if err != nil {
			return maskAny(err)
		}
	}

	// TODO retry operations

	return nil
}

func (c controller) GetStatus(req Request) ([]fleet.UnitStatus, error) {
	extended, err := req.ExtendSlices()
	if err != nil {
		return []fleet.UnitStatus{}, maskAny(err)
	}

	list := []fleet.UnitStatus{}
	for _, unit := range extended.Units {
		status, err := c.Fleet.GetStatus(unit.Name)
		if err != nil {
			return []fleet.UnitStatus{}, maskAny(err)
		}

		list = append(list, status)
	}

	// TODO retry operations

	return list, nil
}
