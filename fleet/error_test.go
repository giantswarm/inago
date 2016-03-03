package fleet

import (
	"fmt"
	"testing"
)

func Test_Fleet_maskAnyf(t *testing.T) {
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
			t.Fatalf("test case %d: output '%s' != expected '%s'", i+1, output, testCase.Expected)
		}
	}
}

func Test_Fleet_errors(t *testing.T) {
	testCases := []struct {
		Output   bool
		Expected bool
	}{
		{
			Output:   IsIPNotFound(ipNotFoundError),
			Expected: true,
		},
		{
			Output:   IsIPNotFound(unitNotFoundError),
			Expected: false,
		},
		{
			Output:   IsUnitNotFound(unitNotFoundError),
			Expected: true,
		},
		{
			Output:   IsUnitNotFound(ipNotFoundError),
			Expected: false,
		},
		{
			Output:   IsInvalidUnitStatus(invalidUnitStatusError),
			Expected: true,
		},
		{
			Output:   IsInvalidUnitStatus(ipNotFoundError),
			Expected: false,
		},
		{
			Output:   IsInvalidEndpoint(invalidEndpointError),
			Expected: true,
		},
		{
			Output:   IsInvalidEndpoint(invalidUnitStatusError),
			Expected: false,
		},
	}

	for i, testCase := range testCases {
		if testCase.Expected != testCase.Output {
			t.Fatalf("test case %d: expected %t got %t", i+1, testCase.Expected, testCase.Output)
		}
	}
}
