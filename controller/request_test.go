package controller

//import (
//	"fmt"
//	"net"
//	"os"
//	"testing"
//
//	"github.com/juju/errgo"
//	. "github.com/onsi/gomega"
//
//	"github.com/giantswarm/inago/controller"
//	"github.com/giantswarm/inago/file-system/fake"
//	"github.com/giantswarm/inago/fleet"
//)
//
//type testFileSystemSetup struct {
//	FileName    string
//	FileContent []byte
//	FilePerm    os.FileMode
//}
//
//func Test_Common_createRequestWithContent(t *testing.T) {
//	testCases := []struct {
//		Setup    []testFileSystemSetup
//		Group    []string
//		Error    error
//		Expected controller.Request
//	}{
//		// This test ensures that loading a single unit from a directory results in
//		// the expected controller request.
//		{
//			Setup: []testFileSystemSetup{
//				{
//					FileName:    "dirname/dirname_unit.service",
//					FileContent: []byte("some unit content"),
//					FilePerm:    os.FileMode(0644),
//				},
//			},
//			Group: []string{"dirname"},
//			Error: nil,
//			Expected: controller.Request{
//				RequestConfig: controller.RequestConfig{
//					SliceIDs: []string{},
//				},
//				Units: []controller.Unit{
//					{
//						Name:    "dirname_unit.service",
//						Content: "some unit content",
//					},
//				},
//			},
//		},
//
//		// This test ensures that trying to load unit files with invalid input
//		// throws an error.
//		{
//			Setup:    []testFileSystemSetup{},
//			Group:    []string{},
//			Error:    invalidArgumentsError,
//			Expected: controller.Request{},
//		},
//
//		// This test ensures that trying to load unit files when no files are in
//		// the file system throws an error.
//		{
//			Setup: []testFileSystemSetup{},
//			Group: []string{"dirname"},
//			Error: &os.PathError{
//				Op:   "open",
//				Path: "dirname",
//				Err:  errgo.New("no such file or directory"),
//			},
//			Expected: controller.Request{},
//		},
//
//		// This test ensures that loading a single unit from a directory with the
//		// slice expression "@1" results in the expected controller request.
//		{
//			Setup: []testFileSystemSetup{
//				{
//					FileName:    "dirname/dirname_unit@.service",
//					FileContent: []byte("some unit content"),
//					FilePerm:    os.FileMode(0644),
//				},
//			},
//			Group: []string{"dirname@1"},
//			Error: nil,
//			Expected: controller.Request{
//				RequestConfig: controller.RequestConfig{
//					SliceIDs: []string{"1"},
//				},
//				Units: []controller.Unit{
//					{
//						Name:    "dirname_unit@.service",
//						Content: "some unit content",
//					},
//				},
//			},
//		},
//
//		// This test ensures that loading a single unit from a directory with the
//		// slice expression "@1", "@foo" and "@5" results in the expected
//		// controller request.
//		{
//			Setup: []testFileSystemSetup{
//				{
//					FileName:    "dirname/dirname_unit@.service",
//					FileContent: []byte("some unit content"),
//					FilePerm:    os.FileMode(0644),
//				},
//			},
//			Input: []string{"dirname@1", "dirname@foo", "dirname@5"},
//			Error: nil,
//			Expected: controller.Request{
//				RequestConfig: controller.RequestConfig{
//					SliceIDs: []string{"1", "foo", "5"},
//				},
//				Units: []controller.Unit{
//					{
//						Name:    "dirname_unit@.service",
//						Content: "some unit content",
//					},
//				},
//			},
//		},
//	}
//
//	for i, testCase := range testCases {
//		newFileSystem := filesystemfake.NewFileSystem()
//
//		for _, setup := range testCase.Setup {
//			err := newFileSystem.WriteFile(setup.FileName, setup.FileContent, setup.FilePerm)
//			if err != nil {
//				t.Fatalf("FileSystem.WriteFile returned error: %#v", err)
//			}
//		}
//
//		newControllerConfig := controller.DefaultConfig()
//		newControllerConfig.FileSystem = newFileSystem
//		newController := controller.NewController(newControllerConfig)
//
//		newRequestConfig := controller.DefaultRequestConfig()
//		newRequestConfig.Group = group
//		newRequestConfig.SliceIDs = strings.Split(strings.Repeat("x", scale), "")
//		req := controller.NewRequest(newRequestConfig)
//
//		req, err := newController.ExtendWithContent(req)
//
//		output, err := createRequestWithContent(testCase.Input)
//		fmt.Printf("output: %#v\n", output)
//		fmt.Printf("err: %#v\n", err)
//		if testCase.Error != nil && err.Error() != testCase.Error.Error() {
//			t.Fatalf("(test case %d) createRequestWithContent was expected to return error: %#v", i+1, testCase.Error)
//		}
//
//		if len(output.SliceIDs) != len(testCase.Expected.SliceIDs) {
//			t.Fatalf("(test case %d) sliceIDs of generated output differs from expected sliceIDs", i+1)
//		}
//
//		for i, outputUnit := range output.Units {
//			if outputUnit.Name != testCase.Expected.Units[i].Name {
//				t.Fatalf("output unit name '%s' is not equal to expected unit name '%s'", outputUnit.Name, testCase.Expected.Units[i].Name)
//			}
//		}
//	}
//}
//
//func Test_Common_createRequest(t *testing.T) {
//	testCases := []struct {
//		Input    []string
//		Error    error
//		Expected controller.Request
//	}{
//		// This test ensures that loading a single unit from a directory results in
//		// the expected controller request.
//		{
//			Input: []string{"dirname"},
//			Error: nil,
//			Expected: controller.Request{
//				SliceIDs: []string{},
//				Units: []controller.Unit{
//					{
//						Name:    "dirname_unit.service",
//						Content: "some unit content",
//					},
//				},
//			},
//		},
//
//		// This test ensures that trying to load unit files with invalid input
//		// throws an error.
//		{
//			Input:    []string{},
//			Error:    invalidArgumentsError,
//			Expected: controller.Request{},
//		},
//
//		// This test ensures that loading a single unit from a directory with the
//		// slice expression "@1" results in the expected controller request.
//		{
//			Input: []string{"dirname@1"},
//			Error: nil,
//			Expected: controller.Request{
//				SliceIDs: []string{"1"},
//				Units: []controller.Unit{
//					{
//						Name:    "dirname_unit@.service",
//						Content: "some unit content",
//					},
//				},
//			},
//		},
//
//		// This test ensures that loading a single unit from a directory with the
//		// slice expression "@1", "@foo" and "@5" results in the expected
//		// controller request.
//		{
//			Input: []string{"dirname@1", "dirname@foo", "dirname@5"},
//			Expected: controller.Request{
//				SliceIDs: []string{"1", "foo", "5"},
//				Units: []controller.Unit{
//					{
//						Name:    "dirname_unit@.service",
//						Content: "some unit content",
//					},
//				},
//			},
//		},
//	}
//
//	for i, testCase := range testCases {
//		output, err := createRequest(testCase.Input)
//		if testCase.Error != nil && err.Error() != testCase.Error.Error() {
//			t.Fatalf("createRequest was expected to return error: %#v", testCase.Error)
//		}
//
//		if len(output.SliceIDs) != len(testCase.Expected.SliceIDs) {
//			t.Fatalf("(test case %d) sliceIDs of generated output differs from expected sliceIDs", i+1)
//		}
//
//		for i, outputUnit := range output.Units {
//			if outputUnit.Name != testCase.Expected.Units[i].Name {
//				t.Fatalf("output unit name '%s' is not equal to expected unit name '%s'", outputUnit.Name, testCase.Expected.Units[i].Name)
//			}
//		}
//	}
//}
