package validator

import (
	"testing"

	"github.com/giantswarm/inago/controller"
)

// TestValidateRequest tests the ValidateRequest function.
func TestValidateRequest(t *testing.T) {
	var tests = []struct {
		request controller.Request
		valid   bool
		err     error
	}{
		{
			request: controller.Request{
				Group: "empty",
			},
			valid: false,
			err:   noUnitsInGroupError,
		},
		{
			request: controller.Request{
				Group: "single",
				Units: []controller.Unit{
					controller.Unit{
						Name: "single-unit.service",
					},
				},
			},
			valid: true,
			err:   nil,
		},
		{
			request: controller.Request{
				Group: "single",
				Units: []controller.Unit{
					controller.Unit{
						Name: "single-unit.service",
					},
					controller.Unit{
						Name: "single-unit2.timer",
					},
				},
			},
			valid: true,
			err:   nil,
		},
		{
			request: controller.Request{
				Group: "scalable",
				Units: []controller.Unit{
					controller.Unit{
						Name: "scalable-unit@.service",
					},
				},
			},
			valid: true,
			err:   nil,
		},
		{
			request: controller.Request{
				Group: "scalable",
				Units: []controller.Unit{
					controller.Unit{
						Name: "scalable-unit@.service",
					},
					controller.Unit{
						Name: "scalable-unit2@.timer",
					},
				},
			},
			valid: true,
			err:   nil,
		},
		{
			request: controller.Request{
				Group: "mix",
				Units: []controller.Unit{
					controller.Unit{
						Name: "mix-unit1.service",
					},
					controller.Unit{
						Name: "mix-unit2@.service",
					},
				},
			},
			valid: false,
			err:   mixedSliceInstanceError,
		},
		{
			request: controller.Request{
				Group: "single",
				Units: []controller.Unit{
					controller.Unit{
						Name: "bad-prefix.service",
					},
				},
			},
			valid: false,
			err:   badUnitPrefixError,
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
		if !test.valid && test.err != err {
			t.Errorf("%v: Returned err '%v' instead of err '%v'", index, err, test.err)
		}
	}
}

// TestValidateMultipleRequest tests the ValidateMultipleRequest function.
func TestValidateMultipleRequest(t *testing.T) {
	var tests = []struct {
		requests []controller.Request
		valid    bool
		err      error
	}{
		{
			requests: []controller.Request{
				controller.Request{
					Group: "a",
				},
				controller.Request{
					Group: "b",
				},
			},
			valid: true,
			err:   nil,
		},
		{
			requests: []controller.Request{
				controller.Request{
					Group: "bat",
				},
				controller.Request{
					Group: "batman",
				},
			},
			valid: false,
			err:   groupsArePrefixError,
		},
		{
			requests: []controller.Request{
				controller.Request{
					Group: "batwoman",
				},
				controller.Request{
					Group: "batman",
				},
			},
			valid: true,
			err:   nil,
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
		if !test.valid && test.err != err {
			t.Errorf("%v: Returned err '%v' instead of err '%v'", index, err, test.err)
		}
	}
}
