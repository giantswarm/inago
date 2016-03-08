package validator

import (
	"testing"

	"github.com/giantswarm/inago/controller/api"
)

// TestValidateRequest tests the ValidateRequest function.
func TestValidateRequest(t *testing.T) {
	var tests = []struct {
		request      api.Request
		valid        bool
		errAssertion func(error) bool
	}{
		// Test a group with no units in it is not valid.
		{
			request: api.Request{
				Group: "empty",
			},
			valid:        false,
			errAssertion: IsNoUnitsInGroup,
		},
		// Test a group with one well-named unit is valid.
		{
			request: api.Request{
				Group: "single",
				Units: []api.Unit{
					api.Unit{
						Name: "single-unit.service",
					},
				},
			},
			valid:        true,
			errAssertion: nil,
		},
		// Test a group with two well-named units is valid.
		{
			request: api.Request{
				Group: "single",
				Units: []api.Unit{
					api.Unit{
						Name: "single-unit.service",
					},
					api.Unit{
						Name: "single-unit2.timer",
					},
				},
			},
			valid:        true,
			errAssertion: nil,
		},
		// Test a group with a scalable unit is valid.
		{
			request: api.Request{
				Group: "scalable",
				Units: []api.Unit{
					api.Unit{
						Name: "scalable-unit@.service",
					},
				},
			},
			valid:        true,
			errAssertion: nil,
		},
		// Test a group with two scalable units is valid.
		{
			request: api.Request{
				Group: "scalable",
				Units: []api.Unit{
					api.Unit{
						Name: "scalable-unit@.service",
					},
					api.Unit{
						Name: "scalable-unit2@.timer",
					},
				},
			},
			valid:        true,
			errAssertion: nil,
		},
		// Test that a group mixing scalable and unscalable units is not valid.
		{
			request: api.Request{
				Group: "mix",
				Units: []api.Unit{
					api.Unit{
						Name: "mix-unit1.service",
					},
					api.Unit{
						Name: "mix-unit2@.service",
					},
				},
			},
			valid:        false,
			errAssertion: IsMixedSliceInstance,
		},
		// Test that units must be prefixed with their group name.
		{
			request: api.Request{
				Group: "single",
				Units: []api.Unit{
					api.Unit{
						Name: "bad-prefix.service",
					},
				},
			},
			valid:        false,
			errAssertion: IsBadUnitPrefix,
		},
		// Test that group names cannot contain @ symbols.
		{
			request: api.Request{
				Group: "bad@groupname@",
				Units: []api.Unit{
					api.Unit{
						Name: "bad@groupname@.service",
					},
				},
			},
			valid:        false,
			errAssertion: IsAtInGroupNameError,
		},
		// Test that unit names cannot contain multiple @ symbols.
		{
			request: api.Request{
				Group: "group",
				Units: []api.Unit{
					api.Unit{
						Name: "group-un@it@.service",
					},
				},
			},
			valid:        false,
			errAssertion: IsMultipleAtInUnitName,
		},
		// Test that a group cannot have multiple units with the same name.
		{
			request: api.Request{
				Group: "group",
				Units: []api.Unit{
					api.Unit{
						Name: "group-unit1@.service",
					},
					api.Unit{
						Name: "group-unit@.service",
					},
					api.Unit{
						Name: "group-unit2@.service",
					},
					api.Unit{
						Name: "group-unit@.service",
					},
				},
			},
			valid:        false,
			errAssertion: IsUnitsSameName,
		},
	}

	for index, test := range tests {
		valid, err := ValidateRequest(test.request)
		if test.valid != valid {
			t.Errorf("%v: Request validity should be: '%v', was '%v'", index, test.valid, valid)
		}
		if test.valid && err != nil {
			t.Errorf("%v: Request should be valid, but returned err: '%v'", index, err)
		}
		if !test.valid && !test.errAssertion(err) {
			t.Errorf("%v: Request should be invalid, but returned incorrect err '%v'", index, err)
		}
	}
}

// TestValidateMultipleRequest tests the ValidateMultipleRequest function.
func TestValidateMultipleRequest(t *testing.T) {
	var tests = []struct {
		requests     []api.Request
		valid        bool
		errAssertion func(error) bool
	}{
		// Test that two differently named groups are valid.
		{
			requests: []api.Request{
				api.Request{
					Group: "a",
				},
				api.Request{
					Group: "b",
				},
			},
			valid:        true,
			errAssertion: nil,
		},
		// Test that groups which are prefixes of another are invalid.
		{
			requests: []api.Request{
				api.Request{
					Group: "bat",
				},
				api.Request{
					Group: "batman",
				},
			},
			valid:        false,
			errAssertion: IsGroupsArePrefix,
		},
		// Test that the group prefix rule applies to the entire group name.
		{
			requests: []api.Request{
				api.Request{
					Group: "batwoman",
				},
				api.Request{
					Group: "batman",
				},
			},
			valid:        true,
			errAssertion: nil,
		},
		// Test that group names must be unique.
		{
			requests: []api.Request{
				api.Request{
					Group: "joker",
				},
				api.Request{
					Group: "joker",
				},
			},
			valid:        false,
			errAssertion: IsGroupsSameName,
		},
	}

	for index, test := range tests {
		valid, err := ValidateMultipleRequest(test.requests)
		if test.valid != valid {
			t.Errorf("%v: Requests validity should be: '%v', was '%v'", index, test.valid, valid)
		}
		if test.valid && err != nil {
			t.Errorf("%v: Requests should be valid, but returned err: '%v'", index, err)
		}
		if !test.valid && !test.errAssertion(err) {
			t.Errorf("%v: Requests should be invalid, but returned incorrect err '%v'", index, err)
		}
	}
}
