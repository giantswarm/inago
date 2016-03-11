package controller

import (
	"os"
	"testing"

	"github.com/juju/errgo"

	"github.com/giantswarm/inago/file-system/fake"
)

type testFileSystemSetup struct {
	FileName    string
	FileContent []byte
	FilePerm    os.FileMode
}

func Test_Request_ExtendWithContent(t *testing.T) {
	testCases := []struct {
		Setup    []testFileSystemSetup
		Error    error
		Input    Request
		Expected Request
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
			Input: Request{
				RequestConfig: RequestConfig{
					Group:    "dirname",
					SliceIDs: []string{},
				},
			},
			Expected: Request{
				RequestConfig: RequestConfig{
					SliceIDs: []string{},
				},
				Units: []Unit{
					{
						Name:    "dirname_unit.service",
						Content: "some unit content",
					},
				},
			},
		},

		// This test ensures that extending an empty request does not inject
		// unwanted files.
		{
			Setup:    []testFileSystemSetup{},
			Error:    nil,
			Input:    Request{},
			Expected: Request{},
		},

		// This test ensures that trying to load unit files when no files are in
		// the file system throws an error.
		{
			Setup: []testFileSystemSetup{},
			Error: &os.PathError{
				Op:   "open",
				Path: "dirname",
				Err:  errgo.New("no such file or directory"),
			},
			Input: Request{
				RequestConfig: RequestConfig{
					Group:    "dirname",
					SliceIDs: []string{},
				},
			},
			Expected: Request{},
		},
	}

	for i, testCase := range testCases {
		newFileSystem := filesystemfake.NewFileSystem()

		for _, setup := range testCase.Setup {
			err := newFileSystem.WriteFile(setup.FileName, setup.FileContent, setup.FilePerm)
			if err != nil {
				t.Fatal("case", i+1, "expected", nil, "got", err)
			}
		}

		newControllerConfig := DefaultConfig()
		newControllerConfig.FileSystem = newFileSystem
		newController := NewController(newControllerConfig)

		output, err := newController.ExtendWithContent(testCase.Input)
		if testCase.Error != nil && err.Error() != testCase.Error.Error() {
			t.Fatal("case", i+1, "expected", testCase.Error, "got", err)
		}

		if len(output.SliceIDs) != len(testCase.Expected.SliceIDs) {
			t.Fatal("case", i+1, "expected", len(testCase.Expected.SliceIDs), "got", len(output.SliceIDs))
		}

		for i, outputUnit := range output.Units {
			if outputUnit.Name != testCase.Expected.Units[i].Name {
				t.Fatal("case", i+1, "expected", testCase.Expected.Units[i].Name, "got", outputUnit.Name)
			}
		}
	}
}
