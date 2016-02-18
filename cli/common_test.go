package cli

import (
	"os"
	"testing"

	"github.com/giantswarm/formica/controller"
	"github.com/giantswarm/formica/file-system/fake"
)

type testFileSystemSetup struct {
	FileName    string
	FileContent []byte
	FilePerm    os.FileMode
}

func Test_Common_createRequestWithContent(t *testing.T) {
	testCases := []struct {
		Setup    []testFileSystemSetup
		Input    []string
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
			Input: []string{"dirname"},
			Expected: controller.Request{
				SliceIDs: []string{},
				Units: []controller.Unit{
					{
						Name:    "dirname_unit.service",
						Content: "some unit content",
					},
				},
			},
		},

		// This test ensures that loading a single unit from a directory with the
		// slice expression "@1" results in the expected controller request.
		{
			Setup: []testFileSystemSetup{
				{
					FileName:    "dirname/dirname_unit@.service",
					FileContent: []byte("some unit content"),
					FilePerm:    os.FileMode(0644),
				},
			},
			Input: []string{"dirname@1"},
			Expected: controller.Request{
				SliceIDs: []string{"1"},
				Units: []controller.Unit{
					{
						Name:    "dirname_unit@.service",
						Content: "some unit content",
					},
				},
			},
		},

		// This test ensures that loading a single unit from a directory with the
		// slice expression "@1", "@foo" and "@5" results in the expected
		// controller request.
		{
			Setup: []testFileSystemSetup{
				{
					FileName:    "dirname/dirname_unit@.service",
					FileContent: []byte("some unit content"),
					FilePerm:    os.FileMode(0644),
				},
			},
			Input: []string{"dirname@1", "dirname@foo", "dirname@5"},
			Expected: controller.Request{
				SliceIDs: []string{"1", "foo", "5"},
				Units: []controller.Unit{
					{
						Name:    "dirname_unit@.service",
						Content: "some unit content",
					},
				},
			},
		},
	}

	for i, testCase := range testCases {
		newFileSystem = filesystemfake.NewFileSystem()

		for _, setup := range testCase.Setup {
			err := newFileSystem.WriteFile(setup.FileName, setup.FileContent, setup.FilePerm)
			if err != nil {
				t.Fatalf("FileSystem.WriteFile returned error: %#v", err)
			}
		}

		output, err := createRequestWithContent(testCase.Input)
		if err != nil {
			t.Fatalf("createRequestWithContent returned error: %#v", err)
		}

		if len(output.SliceIDs) != len(testCase.Expected.SliceIDs) {
			t.Fatalf("(test case %d) sliceIDs of generated output differs from expected sliceIDs", i+1)
		}

		for i, outputUnit := range output.Units {
			if outputUnit.Name != testCase.Expected.Units[i].Name {
				t.Fatalf("output unit name '%s' is not equal to expected unit name '%s'", outputUnit.Name, testCase.Expected.Units[i].Name)
			}
		}
	}
}

func Test_Common_createRequest(t *testing.T) {
	testCases := []struct {
		Input    []string
		Expected controller.Request
	}{
		// This test ensures that loading a single unit from a directory results in
		// the expected controller request.
		{
			Input: []string{"dirname"},
			Expected: controller.Request{
				SliceIDs: []string{},
				Units: []controller.Unit{
					{
						Name:    "dirname_unit.service",
						Content: "some unit content",
					},
				},
			},
		},

		// This test ensures that loading a single unit from a directory with the
		// slice expression "@1" results in the expected controller request.
		{
			Input: []string{"dirname@1"},
			Expected: controller.Request{
				SliceIDs: []string{"1"},
				Units: []controller.Unit{
					{
						Name:    "dirname_unit@.service",
						Content: "some unit content",
					},
				},
			},
		},

		// This test ensures that loading a single unit from a directory with the
		// slice expression "@1", "@foo" and "@5" results in the expected
		// controller request.
		{
			Input: []string{"dirname@1", "dirname@foo", "dirname@5"},
			Expected: controller.Request{
				SliceIDs: []string{"1", "foo", "5"},
				Units: []controller.Unit{
					{
						Name:    "dirname_unit@.service",
						Content: "some unit content",
					},
				},
			},
		},
	}

	for i, testCase := range testCases {
		output, err := createRequest(testCase.Input)
		if err != nil {
			t.Fatalf("createRequestWithContent returned error: %#v", err)
		}

		if len(output.SliceIDs) != len(testCase.Expected.SliceIDs) {
			t.Fatalf("(test case %d) sliceIDs of generated output differs from expected sliceIDs", i+1)
		}

		for i, outputUnit := range output.Units {
			if outputUnit.Name != testCase.Expected.Units[i].Name {
				t.Fatalf("output unit name '%s' is not equal to expected unit name '%s'", outputUnit.Name, testCase.Expected.Units[i].Name)
			}
		}
	}
}
