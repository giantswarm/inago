package controller

import (
	"fmt"
	"testing"
)

func Test_Controller_maskAnyf(t *testing.T) {
	testCases := []struct {
		InputError  error
		InputFormat string
		InputArgs   []interface{}
		Expected    error
	}{
		{
			InputError:  nil,
			InputFormat: "",
			InputArgs:   []interface{}{},
			Expected:    nil,
		},
		{
			InputError:  fmt.Errorf("foo"),
			InputFormat: "bar",
			InputArgs:   []interface{}{},
			Expected:    nil,
		},
		{
			InputError:  fmt.Errorf("foo"),
			InputFormat: "bar %s",
			InputArgs:   []interface{}{"baz"},
			Expected:    fmt.Errorf("foo: bar baz"),
		},
	}

	for i, testCase := range testCases {
		var output error
		if len(testCase.InputArgs) == 0 {
			output = maskAnyf(testCase.InputError, testCase.InputFormat)
		} else {
			output = maskAnyf(testCase.InputError, testCase.InputFormat, testCase.InputArgs...)
		}

		if testCase.Expected != nil && output.Error() != testCase.Expected.Error() {
			t.Fatalf("test case %d: output '%s' != expected '%s'", i, output, testCase.Expected)
		}
	}
}

func Test_Controller_errors(t *testing.T) {
	testCases := []struct {
		Output   bool
		Expected bool
	}{
		{
			Output:   IsUnitNotFound(unitNotFoundError),
			Expected: true,
		},
		{
			Output:   IsUnitNotFound(unitSliceNotFoundError),
			Expected: false,
		},
		{
			Output:   IsInvalidUnitStatus(invalidUnitStatusError),
			Expected: true,
		},
		{
			Output:   IsInvalidUnitStatus(unitSliceNotFoundError),
			Expected: false,
		},
		{
			Output:   IsInvalidArgument(invalidArgumentError),
			Expected: true,
		},
		{
			Output:   IsInvalidArgument(unitSliceNotFoundError),
			Expected: false,
		},
		{
			Output:   IsWaitTimeoutReached(waitTimeoutReachedError),
			Expected: true,
		},
		{
			Output:   IsWaitTimeoutReached(invalidUnitStatusError),
			Expected: false,
		},
		{
			Output:   IsUnitSliceNotFound(unitSliceNotFoundError),
			Expected: true,
		},
		{
			Output:   IsUnitSliceNotFound(waitTimeoutReachedError),
			Expected: false,
		},
	}

	for i, testCase := range testCases {
		if testCase.Expected != testCase.Output {
			t.Fatalf("test case %d: expected %t got %t", i+1, testCase.Expected, testCase.Output)
		}
	}
}
