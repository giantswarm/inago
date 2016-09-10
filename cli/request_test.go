package cli

import (
	"os"
	"reflect"
	"testing"

	"github.com/juju/errgo"
	"github.com/spf13/afero"

	"github.com/giantswarm/inago/controller"
)

func givenSomeUnitFileContent() string {
	return "[Unit]\n" +
		"Description=Some Unit File Content\n" +
		"\n" +
		"[Service]\n" +
		"ExecStart=/bin/bash -c 'while true; do echo nothing to see, go along; done'\n"

}

type testFileSystemSetup struct {
	FileName    string
	FileContent []byte
	FilePerm    os.FileMode
}

func Test_Request_ExtendWithContent(t *testing.T) {
	testCases := []struct {
		Setup    []testFileSystemSetup
		Error    error
		Input    controller.Request
		Expected controller.Request
	}{
		// This test ensures that loading a single unit from a directory results in
		// the expected controller request.
		{
			Setup: []testFileSystemSetup{
				{
					FileName:    "dirname/dirname_unit.service",
					FileContent: []byte("some unit content"),
					FilePerm:    os.FileMode(0644),
				},
			},
			Error: nil,
			Input: controller.Request{
				RequestConfig: controller.RequestConfig{
					Group:    "dirname",
					SliceIDs: []string{},
				},
			},
			Expected: controller.Request{
				RequestConfig: controller.RequestConfig{
					SliceIDs: []string{},
				},
				Units: []controller.Unit{
					{
						Name:    "dirname_unit.service",
						Content: givenSomeUnitFileContent(),
					},
				},
			},
		},

		// This test ensures that extending an empty request does not inject
		// unwanted files.
		{
			Setup:    []testFileSystemSetup{},
			Error:    noUnitFilesError,
			Input:    controller.Request{},
			Expected: controller.Request{},
		},

		// This test ensures that trying to load unit files when no files are in
		// the file system throws an error.
		{
			Setup: []testFileSystemSetup{},
			Error: &os.PathError{
				Op:   "open",
				Path: "dirname",
				Err:  os.ErrNotExist,
			},
			Input: controller.Request{
				RequestConfig: controller.RequestConfig{
					Group:    "dirname",
					SliceIDs: []string{},
				},
			},
			Expected: controller.Request{},
		},

		// This test ensures that folders inside a group folder are ignored
		{
			Setup: []testFileSystemSetup{
				{FileName: "groupname/someotherdiretctory/REAMDE.md", FileContent: []byte("DO NOT READ ME"), FilePerm: os.FileMode(0644)},
				{FileName: "groupname/groupname-1.service", FileContent: []byte(givenSomeUnitFileContent()), FilePerm: os.FileMode(0644)},
				{FileName: "groupname/groupname-2.service", FileContent: []byte(givenSomeUnitFileContent()), FilePerm: os.FileMode(0644)},
			},
			Input: controller.Request{
				RequestConfig: controller.RequestConfig{
					Group: "groupname",
				},
			},
			Expected: controller.Request{
				Units: []controller.Unit{
					{
						Name:    "groupname-1.service",
						Content: givenSomeUnitFileContent(),
					},
					{
						Name:    "groupname-2.service",
						Content: givenSomeUnitFileContent(),
					},
				},
			},
		},
	}

	for i, testCase := range testCases {
		newFileSystem := afero.Afero{afero.NewMemMapFs()}

		for _, setup := range testCase.Setup {
			err := newFileSystem.WriteFile(setup.FileName, setup.FileContent, setup.FilePerm)
			if err != nil {
				t.Fatal("case", i+1, "expected", nil, "got", err)
			}
		}

		output, err := extendRequestWithContent(newFileSystem, testCase.Input)
		if !reflect.DeepEqual(errgo.Cause(err), errgo.Cause(testCase.Error)) {
			t.Fatalf("case %d: expected %v, got %v", i+1, testCase.Error, err)
		}

		if len(output.SliceIDs) != len(testCase.Expected.SliceIDs) {
			t.Fatal("case", i+1, "expected", len(testCase.Expected.SliceIDs), "got", len(output.SliceIDs))
		}

		if len(output.Units) != len(testCase.Expected.Units) {
			t.Fatalf("case %d: expected %d units in output, got %d", i+1, len(testCase.Expected.Units), len(output.Units))
		}
		for _, outputUnit := range testCase.Expected.Units {
			found := false
			for _, expectedUnit := range output.Units {
				if outputUnit.Name == expectedUnit.Name {
					found = true
				}
			}
			if !found {
				t.Fatalf("case %d: expected %s to be in output, not found", i+1, outputUnit.Name)
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
