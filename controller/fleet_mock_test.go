package controller

import (
	"fmt"
	"regexp"

	"github.com/stretchr/testify/mock"

	"github.com/giantswarm/inago/fleet"
)

type fleetMockConfig struct {
	UseTestifyMock        bool
	UseCustomMock         bool
	CustomMockUsed        int
	FirstCustomMockStatus []fleet.UnitStatus
	LastCustomMockStatus  []fleet.UnitStatus
}

func defaultFleetMockConfig() fleetMockConfig {
	newConfig := fleetMockConfig{
		UseTestifyMock:        true,
		UseCustomMock:         false,
		CustomMockUsed:        0,
		FirstCustomMockStatus: nil,
		LastCustomMockStatus:  nil,
	}

	return newConfig
}

func newFleetMock(config fleetMockConfig) *fleetMock {
	newMock := &fleetMock{
		fleetMockConfig: config,
	}

	return newMock
}

type fleetMock struct {
	fleetMockConfig
	mock.Mock
}

func (fm *fleetMock) Submit(name, content string) error {
	args := fm.Called(name, content)
	return args.Error(0)
}
func (fm *fleetMock) Start(name string) error {
	args := fm.Called(name)
	return args.Error(0)
}
func (fm *fleetMock) Stop(name string) error {
	args := fm.Called(name)
	return args.Error(0)
}
func (fm *fleetMock) Destroy(name string) error {
	args := fm.Called(name)
	return args.Error(0)
}
func (fm *fleetMock) GetStatus(name string) (fleet.UnitStatus, error) {
	args := fm.Called(name)
	return args.Get(0).(fleet.UnitStatus), args.Error(1)
}
func (fm *fleetMock) GetStatusWithExpression(exp *regexp.Regexp) ([]fleet.UnitStatus, error) {
	args := fm.Called(exp)
	return args.Get(0).([]fleet.UnitStatus), args.Error(1)
}
func (fm *fleetMock) GetStatusWithMatcher(f func(string) bool) ([]fleet.UnitStatus, error) {
	if fm.UseTestifyMock {
		args := fm.Called(f)
		return args.Get(0).([]fleet.UnitStatus), args.Error(1)
	} else if fm.UseCustomMock {
		fm.CustomMockUsed++
		if fm.CustomMockUsed <= 3 {
			return fm.FirstCustomMockStatus, nil
		} else {
			return fm.LastCustomMockStatus, nil
		}
	}

	return nil, fmt.Errorf("invalid mock setup")
}
