package controller

import (
	"fmt"
	"regexp"

	"github.com/giantswarm/inago/fleet"
)

// DefaultRequestConfig returns a RequestConfig by best effort.
func DefaultRequestConfig() RequestConfig {
	newConfig := RequestConfig{
		Group:    "",
		SliceIDs: []string{},
	}

	return newConfig
}

// RequestConfig represents the configuration of a Request.
type RequestConfig struct {
	// Group represents the plain group name without any slice expression.
	Group string

	// SliceIDs contains the IDs to create. IDs can be "1", "first", "whatever",
	// "5", etc..
	SliceIDs []string
}

// Request represents a controller request. This is used to process some action
// on the controller.
type Request struct {
	RequestConfig

	// Units represents a list of unit files that is supposed to be extended
	// using the provided slice IDs.
	Units []Unit

	// DesiredSlices defines the number of random sliceIDs that should be generated
	// when submitting new groups.
	DesiredSlices int
}

// NewRequest returns a Request, given a RequestConfig.
func NewRequest(config RequestConfig) Request {
	req := Request{
		RequestConfig: config,
		Units:         []Unit{},
	}

	return req
}

var unitExp = regexp.MustCompile("@.")

// isSliceable checks whether all units of the request are sliceable (contain an @)
func (r Request) isSliceable() bool {
	for _, unit := range r.Units {
		if !unitExp.MatchString(unit.Name) {
			return false
		}
	}
	return true
}

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
	if len(r.SliceIDs) == 0 {
		return r, nil
	}

	var newUnits []Unit
	for _, sliceID := range r.SliceIDs {
		for _, unit := range r.Units {
			newUnit := unit
			// TODO fix extension
			newUnit.Name = unitExp.ReplaceAllString(newUnit.Name, fmt.Sprintf("@%s.", sliceID))
			newUnits = append(newUnits, newUnit)
		}
	}
	r.Units = newUnits

	return r, nil
}

func (c controller) getExistingSliceIDs(req Request) ([]string, error) {
	usl, err := c.Fleet.GetStatusWithMatcher(matchesUnitBase(req))
	if fleet.IsUnitNotFound(err) {
		// This happenes when there is no unit, e.g. on submit. Thus we don't need
		// to check against anything. Se we do nothing and go ahead by simply
		// creating a new random ID.
	} else if err != nil {
		return nil, maskAny(err)
	}

	var newSliceIDs []string
	for _, us := range usl {
		if us.SliceID == "" {
			// This unit has no explicit slice ID. Skip it.
			continue
		}
		if contains(newSliceIDs, us.SliceID) {
			// We already tracked this ID. Go ahead.
			continue
		}
		newSliceIDs = append(newSliceIDs, us.SliceID)
	}

	return newSliceIDs, nil
}

func (c controller) ExtendWithExistingSliceIDs(req Request) (Request, error) {
	newSliceIDs, err := c.getExistingSliceIDs(req)
	if err != nil {
		return Request{}, maskAny(err)
	}
	req.SliceIDs = newSliceIDs

	return req, nil
}

func contains(l []string, e string) bool {
	for _, le := range l {
		if le == e {
			return true
		}
	}

	return false
}

func (c controller) ExtendWithRandomSliceIDs(req Request) (Request, error) {
	if !req.isSliceable() {
		return req, nil
	}

	// Lookup existing slice IDs.
	usl, err := c.groupStatusWithValidate(req)
	if IsUnitNotFound(err) {
		// This happens when no unit is found, e.g. on submit. In this case we
		// simply go ahead, because we have no existing IDs to ignore.
	} else if err != nil {
		return Request{}, maskAny(err)
	}

	// Find enough sufficient IDs.
	var newIDs []string
	for i := 0; i < req.DesiredSlices; i++ {
		for {
			newID := NewID()

			ok, err := containsUnitStatusSliceID(usl, newID)
			if err != nil {
				return Request{}, maskAny(err)
			}
			if ok {
				// We already have this ID in the group. Try again.
				continue
			}
			if contains(newIDs, newID) {
				// We already created this ID. Try again.
				continue
			}

			newIDs = append(newIDs, newID)
			break
		}
	}
	req.SliceIDs = newIDs
	req.DesiredSlices = 0

	return req, nil
}
