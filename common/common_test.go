package common

import (
	"fmt"
	"testing"

	"github.com/juju/errgo"
)

func Test_SliceID(t *testing.T) {
	var testCases = []struct {
		Input    string
		Expected string
		Error    error
	}{
		{
			Input:    "app@1.service",
			Expected: "1",
			Error:    nil,
		},
		{
			Input:    "app@foo.service",
			Expected: "foo",
			Error:    nil,
		},
		{
			Input:    "app@1.mount",
			Expected: "1",
			Error:    nil,
		},
		{
			Input:    "app@foo.mount",
			Expected: "foo",
			Error:    nil,
		},

		{
			Input:    "app.service",
			Expected: "",
			Error:    nil,
		},
		{
			Input:    "app.mount",
			Expected: "",
			Error:    nil,
		},

		{
			Input:    "app@1",
			Expected: "1",
			Error:    nil,
		},
		{
			Input:    "app@foo",
			Expected: "foo",
			Error:    nil,
		},

		{
			Input: `
				app@foo
				app@foo
			`,
			Expected: "",
			Error:    invalidArgumentsError,
		},
	}

	for i, testCase := range testCases {
		output, err := SliceID(testCase.Input)
		if errgo.Cause(err) != errgo.Cause(testCase.Error) {
			t.Fatal("case", i+1, "expected", errgo.Cause(testCase.Error), "got", errgo.Cause(err))
		}
		if output != testCase.Expected {
			t.Fatal("case", i+1, "expected", testCase.Expected, "got", output)
		}
	}
}

func Test_UnitBase(t *testing.T) {
	var testCases = []struct {
		Input    string
		Expected string
	}{
		{
			Input:    "app@1.service",
			Expected: "app",
		},
		{
			Input:    "app@foo.service",
			Expected: "app",
		},
		{
			Input:    "app@1.mount",
			Expected: "app",
		},
		{
			Input:    "app@foo.mount",
			Expected: "app",
		},

		{
			Input:    "app.service",
			Expected: "app",
		},
		{
			Input:    "app.mount",
			Expected: "app",
		},

		{
			Input:    "app@1",
			Expected: "app",
		},
		{
			Input:    "app@foo",
			Expected: "app",
		},
	}

	for i, testCase := range testCases {
		output := UnitBase(testCase.Input)
		if output != testCase.Expected {
			t.Fatal("case", i+1, "expected", testCase.Expected, "got", output)
		}
	}
}

// ExampleSliceID is an example of the SliceID function.
func ExampleSliceID() {
	for _, input := range []string{
		"app@1.service",
		"app@1.mount",
		"app.service",
		"app.mount",
	} {
		id, _ := SliceID(input)
		fmt.Println(id)
	}

	// Output: 1
	// 1
	//
	//
}

// ExampleUnitBase is an example of the UnitBase function.
func ExampleUnitBase() {
	for _, input := range []string{
		"app@1.service",
		"app@1.mount",
		"app.service",
		"app.mount",
	} {
		fmt.Println(UnitBase(input))
	}

	// Output: app
	// app
	// app
	// app
}
