package cli

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/juju/errgo"

	"github.com/giantswarm/inago/controller"
)

type fileDesc struct {
	Path    string
	Content string
}

func TestNewRequestWithUnits(t *testing.T) {
	tests := []struct {
		GroupDir string
		Files    []fileDesc
		Error    error
		Group    string
		Units    []controller.Unit
	}{
		{
			GroupDir: "g1",
			Files: []fileDesc{
				{Path: "g1/g1-1.service", Content: "unit1"},
			},
			Error: nil,
			Group: "g1",
			Units: []controller.Unit{
				{Name: "g1-1.service", Content: "unit1"},
			},
		},
		{
			GroupDir: "g2",
			Files:    []fileDesc{},
			Error:    noUnitFilesError,
			Group:    "g_ignore",
			Units:    []controller.Unit{},
		},
		{
			GroupDir: "g3",
			Files: []fileDesc{
				{Path: "g3/g3-dir/g3-1.service", Content: "DO NOT READ ME"},
				{Path: "g3/badprefix-1.service", Content: "DO NOT READ ME"},
				{Path: "g3/g3-1.service", Content: "unit1"},
				{Path: "g3/g3-2.service", Content: "unit2"},
			},
			Error: nil,
			Group: "g3",
			Units: []controller.Unit{
				{Name: "g3-1.service", Content: "unit1"},
				{Name: "g3-2.service", Content: "unit2"},
			},
		},
		{
			GroupDir: "a/g4",
			Files: []fileDesc{
				{Path: "a/g4/g4-1.service", Content: "unit1"},
			},
			Error: nil,
			Group: "g4",
			Units: []controller.Unit{
				{Name: "g4-1.service", Content: "unit1"},
			},
		},
		{
			GroupDir: "././/a/b/c/../c/g5/d/..",
			Files: []fileDesc{
				{Path: "a/b/c/g5/g5-1.service", Content: "unit1"},
			},
			Error: nil,
			Group: "g5",
			Units: []controller.Unit{
				{Name: "g5-1.service", Content: "unit1"},
			},
		},
	}

	cleanDir := chdirTmp(t)
	defer cleanDir()

	for i, tt := range tests {
		prepareDir(t, i, filepath.Clean(tt.GroupDir), tt.Files)
		req, err := newRequestWithUnits(tt.GroupDir)

		if !reflect.DeepEqual(errgo.Cause(err), tt.Error) {
			t.Errorf("#%d: unexpected error = %v", i, err)
		}
		if tt.Error != nil {
			continue
		}
		if req.Group != tt.Group {
			t.Errorf("#%d: Group = %s, want %s", i, req.Group, tt.Group)
		}
		if !reflect.DeepEqual(req.Units, tt.Units) {
			t.Errorf("#%d: Units = %s, want %s", i, req.Units, tt.Units)
		}
	}
}

func Test_Request_ParseGroupCLIArgs_Success(t *testing.T) {
	type Expected struct {
		Group    string
		SliceIDs []string
	}
	testCases := []struct {
		Input    []string
		Expected Expected
	}{
		// Tests that no sliceIDs are returned when non where provided
		{
			Input: []string{"mygroup"},
			Expected: Expected{
				Group:    "mygroup",
				SliceIDs: []string{},
			},
		},
		// Tests that group and slice are split correctly
		{
			Input: []string{"mygroup@1"},
			Expected: Expected{
				Group:    "mygroup",
				SliceIDs: []string{"1"},
			},
		},
		// Tests that multiple group sliceIDs are split correctly
		{
			Input: []string{"mygroup@1", "mygroup@2"},
			Expected: Expected{
				Group:    "mygroup",
				SliceIDs: []string{"1", "2"},
			},
		},
	}

	for _, test := range testCases {
		group, sliceIDs, err := parseGroupCLIArgs(test.Input)
		if err != nil {
			t.Fatalf("got unexpected error: %v", err)
		}

		if group != test.Expected.Group {
			t.Fatalf("got group %v, expected group to be %v.", group, test.Expected.Group)
		}
		if !reflect.DeepEqual(sliceIDs, test.Expected.SliceIDs) {
			t.Fatalf("got sliceIDs %v, expected sliceIDs to be %v.", sliceIDs, test.Expected.SliceIDs)
		}
	}
}

func Test_Request_ParseGroupCLIArgs_Error(t *testing.T) {
	testCases := []struct {
		Input      []string
		CheckError func(error) bool
	}{ // Tests that mixed groups with sliceIDs return an invalidArgumentsError
		{
			Input:      []string{"mygroup@1", "othergroup@2"},
			CheckError: IsInvalidArgumentsError,
		},
		// Tests that using two different groups fails
		{
			Input:      []string{"mygroup", "othergroup"},
			CheckError: IsInvalidArgumentsError,
		},
	}

	for _, test := range testCases {
		_, _, err := parseGroupCLIArgs(test.Input)
		if !test.CheckError(err) {
			t.Fatalf("got unexpected Error '%v'", err)
		}
	}
}
