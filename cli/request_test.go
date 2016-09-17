package cli

import (
	"reflect"
	"testing"

	"github.com/juju/errgo"

	"github.com/giantswarm/inago/controller"
)

type fileDesc struct {
	Path    string
	Content string
}

func Test_Request_ExtendWithContent(t *testing.T) {
	testCases := []struct {
		Group string
		Files []fileDesc
		Units []controller.Unit
		Error error
	}{
		// This test ensures that loading a single unit from a directory results in
		// the expected controller request.
		{
			Group: "dirname",
			Files: []fileDesc{
				{Path: "dirname/dirname_unit.service", Content: "unit1"},
			},
			Units: []controller.Unit{
				{Name: "dirname_unit.service", Content: "unit1"},
			},
			Error: nil,
		},

		// This test ensures that noUnitFilesError is returned for empty gorup.
		{
			Group: "",
			Files: []fileDesc{},
			Units: []controller.Unit{},
			Error: groupNotExistError,
		},

		// This test ensures that trying to load unit files when no files are in
		// the file system throws an error.
		{
			Group: "g2",
			Files: []fileDesc{},
			Units: []controller.Unit{},
			Error: noUnitFilesError,
		},

		// This test ensures that: all directories inside a group directory are ignored;
		// regular files inside group directory not prefixed with group name are ignored.
		{
			Group: "groupname",
			Files: []fileDesc{
				{Path: "groupname/gorupname-dir/REAMDE.md", Content: "DO NOT READ ME"},
				{Path: "groupname/badprefix-1.service", Content: "DO NOT READ ME"},
				{Path: "groupname/groupname-1.service", Content: "unit1"},
				{Path: "groupname/groupname-2.service", Content: "unit2"},
			},
			Units: []controller.Unit{
				{Name: "groupname-1.service", Content: "unit1"},
				{Name: "groupname-2.service", Content: "unit2"},
			},
			Error: nil,
		},
	}

	cleanDir := chdirTmp(t)
	defer cleanDir()

	for i, testCase := range testCases {
		prepareDir(t, i, testCase.Group, testCase.Files)
		req := controller.NewRequest(controller.RequestConfig{Group: testCase.Group})
		req, err := extendRequestWithContent(req)

		if !reflect.DeepEqual(errgo.Cause(err), testCase.Error) {
			t.Fatalf("case %d: expected %v, got %v", i+1, testCase.Error, errgo.Cause(err))
		}

		if len(req.Units) != len(testCase.Units) {
			t.Errorf("case %d: expected %d, got %d", i+1, len(testCase.Units), len(req.Units))
		}
		for _, wunit := range testCase.Units {
			found := false
			for _, unit := range req.Units {
				if reflect.DeepEqual(unit, wunit) {
					found = true
				}
			}
			if !found {
				t.Errorf("case %d: expected %v, got none", i+1, wunit)
			}
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
