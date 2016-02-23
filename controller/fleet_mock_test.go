package controller

import (
	"regexp"

	"github.com/stretchr/testify/mock"

	"github.com/giantswarm/formica/fleet"
)

type fleetMock struct {
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
	args := fm.Called(f)
	return args.Get(0).([]fleet.UnitStatus), args.Error(1)
}
