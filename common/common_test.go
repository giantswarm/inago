package common

import (
	"testing"
)

func Test_SliceID(t *testing.T) {
	var testCases = []struct {
		Input    string
		Expected string
	}{
		{
			Input:    "app@1.service",
			Expected: "1",
		},
		{
			Input:    "app@foo.service",
			Expected: "foo",
		},
		{
			Input:    "app@1.mount",
			Expected: "1",
		},
		{
			Input:    "app@foo.mount",
			Expected: "foo",
		},

		{
			Input:    "app.service",
			Expected: "",
		},
		{
			Input:    "app.mount",
			Expected: "",
		},

		{
			Input:    "app@1",
			Expected: "1",
		},
		{
			Input:    "app@foo",
			Expected: "foo",
		},
	}

	for i, testCase := range testCases {
		output, err := SliceID(testCase.Input)
		if err != nil {
			t.Fatal("case", i+1, "expected", nil, "got", err)
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
