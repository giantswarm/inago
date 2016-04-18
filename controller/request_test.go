package controller

import (
	"reflect"
	"testing"
)

func Test_Request_NewRequest(t *testing.T) {
	testCases := []struct {
		Input    RequestConfig
		Expected Request
	}{
		{
			Input:    RequestConfig{Group: "group", SliceIDs: []string{"1", "2"}},
			Expected: Request{RequestConfig: RequestConfig{Group: "group", SliceIDs: []string{"1", "2"}}, Units: []Unit{}},
		},
		{
			Input:    RequestConfig{Group: "group/", SliceIDs: []string{"1", "2"}},
			Expected: Request{RequestConfig: RequestConfig{Group: "group", SliceIDs: []string{"1", "2"}}, Units: []Unit{}},
		},
		{
			Input:    RequestConfig{Group: "group/", SliceIDs: []string{"1", "2"}},
			Expected: Request{RequestConfig: RequestConfig{Group: "group", SliceIDs: []string{"1", "2"}}, Units: []Unit{}},
		},
		{
			Input:    RequestConfig{Group: "group//", SliceIDs: []string{"1", "2"}},
			Expected: Request{RequestConfig: RequestConfig{Group: "group", SliceIDs: []string{"1", "2"}}, Units: []Unit{}},
		},
		{
			Input:    RequestConfig{Group: "group/////", SliceIDs: []string{"1", "2"}},
			Expected: Request{RequestConfig: RequestConfig{Group: "group", SliceIDs: []string{"1", "2"}}, Units: []Unit{}},
		},
		{
			Input:    RequestConfig{Group: "group/foo////", SliceIDs: []string{"1", "2"}},
			Expected: Request{RequestConfig: RequestConfig{Group: "group/foo", SliceIDs: []string{"1", "2"}}, Units: []Unit{}},
		},
	}

	for i, testCase := range testCases {
		newRequest := NewRequest(testCase.Input)
		if !reflect.DeepEqual(newRequest, testCase.Expected) {
			t.Fatal("case", i, "expected", testCase.Expected, "got", newRequest)
		}
	}
}
