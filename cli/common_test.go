package cli

import (
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/juju/errgo"
	. "github.com/onsi/gomega"

	"github.com/giantswarm/inago/controller"
	"github.com/giantswarm/inago/file-system/fake"
	"github.com/giantswarm/inago/fleet"
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
		Error    error
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
			Error: nil,
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

		// This test ensures that trying to load unit files with invalid input
		// throws an error.
		{
			Setup:    []testFileSystemSetup{},
			Input:    []string{},
			Error:    invalidArgumentsError,
			Expected: controller.Request{},
		},

		// This test ensures that trying to load unit files when no files are in
		// the file system throws an error.
		{
			Setup: []testFileSystemSetup{},
			Input: []string{"dirname"},
			Error: &os.PathError{
				Op:   "open",
				Path: "dirname",
				Err:  errgo.New("no such file or directory"),
			},
			Expected: controller.Request{},
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
			Error: nil,
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
			Error: nil,
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
		fmt.Printf("output: %#v\n", output)
		fmt.Printf("err: %#v\n", err)
		if testCase.Error != nil && err.Error() != testCase.Error.Error() {
			t.Fatalf("(test case %d) createRequestWithContent was expected to return error: %#v", i+1, testCase.Error)
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
		Error    error
		Expected controller.Request
	}{
		// This test ensures that loading a single unit from a directory results in
		// the expected controller request.
		{
			Input: []string{"dirname"},
			Error: nil,
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

		// This test ensures that trying to load unit files with invalid input
		// throws an error.
		{
			Input:    []string{},
			Error:    invalidArgumentsError,
			Expected: controller.Request{},
		},

		// This test ensures that loading a single unit from a directory with the
		// slice expression "@1" results in the expected controller request.
		{
			Input: []string{"dirname@1"},
			Error: nil,
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
		if testCase.Error != nil && err.Error() != testCase.Error.Error() {
			t.Fatalf("createRequest was expected to return error: %#v", testCase.Error)
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

func Test_Common_createStatus(t *testing.T) {
	RegisterTestingT(t)

	type input struct {
		Group   string
		USL     controller.UnitStatusList
		Verbose bool
	}
	type testCase struct {
		Comment  string
		Expected []string
		Input    input
	}

	testCases := []testCase{
		// 3 slices of 1 group with 2 units
		testCase{
			Comment: "3 slices of 1 group with 2 units",
			Expected: []string{
				"Group | Units | FDState | FCState | SAState | IP | Machine",
				"",
				"example@1 | * | loaded | loaded | inactive | 172.17.8.101 | 505e0d7802d7439a924c269b76f34b5f",
				"example@2 | * | loaded | loaded | inactive | 172.17.8.102 | 9ebb53b04b0d46fb94b4fd1b3f562d2b",
				"example@3 | * | loaded | loaded | inactive | 172.17.8.103 | e3cb5f13a9164ba5b7eff6c920475e61",
				"",
			},
			Input: input{
				Group: "example",
				USL: controller.UnitStatusList{
					unitStatus("example-foo@1.service", "@1", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "loaded", "loaded"),
					unitStatus("example-bar@1.service", "@1", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "loaded", "loaded"),
					unitStatus("example-foo@2.service", "@2", "172.17.8.102", "9ebb53b04b0d46fb94b4fd1b3f562d2b", "loaded", "loaded"),
					unitStatus("example-bar@2.service", "@2", "172.17.8.102", "9ebb53b04b0d46fb94b4fd1b3f562d2b", "loaded", "loaded"),
					unitStatus("example-foo@3.service", "@3", "172.17.8.103", "e3cb5f13a9164ba5b7eff6c920475e61", "loaded", "loaded"),
					unitStatus("example-bar@3.service", "@3", "172.17.8.103", "e3cb5f13a9164ba5b7eff6c920475e61", "loaded", "loaded"),
				},
				Verbose: false,
			},
		},
		// 1 slice of 1 group with 2 units
		testCase{
			Comment: "1 slice of 1 group with 2 units",
			Expected: []string{
				"Group | Units | FDState | FCState | SAState | IP | Machine",
				"",
				"example@1 | * | loaded | loaded | inactive | 172.17.8.101 | 505e0d7802d7439a924c269b76f34b5f",
				"",
			},
			Input: input{
				Group: "example",
				USL: controller.UnitStatusList{
					unitStatus("example-foo", "@1", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "loaded", "loaded"),
					unitStatus("example-bar", "@1", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "loaded", "loaded"),
				},
				Verbose: false,
			},
		},
		// 3 slices of 1 group with 2 units - verbose
		testCase{
			Comment: "3 slices of 1 group with 2 units - verbose",
			Expected: []string{
				"Group | Units | FDState | FCState | SAState | Hash | IP | Machine",
				"",
				"example@1 | example-foo@1.service | loaded | loaded | inactive | 4311 | 172.17.8.101 | 505e0d7802d7439a924c269b76f34b5f",
				"example@1 | example-bar@1.service | loaded | loaded | inactive | 4311 | 172.17.8.101 | 505e0d7802d7439a924c269b76f34b5f",
				"example@2 | example-foo@2.service | loaded | loaded | inactive | 4311 | 172.17.8.102 | 9ebb53b04b0d46fb94b4fd1b3f562d2b",
				"example@2 | example-bar@2.service | loaded | loaded | inactive | 4311 | 172.17.8.102 | 9ebb53b04b0d46fb94b4fd1b3f562d2b",
				"example@3 | example-foo@3.service | loaded | loaded | inactive | 4311 | 172.17.8.103 | e3cb5f13a9164ba5b7eff6c920475e61",
				"example@3 | example-bar@3.service | loaded | loaded | inactive | 4311 | 172.17.8.103 | e3cb5f13a9164ba5b7eff6c920475e61",
				"",
			},
			Input: input{
				Group: "example",
				USL: controller.UnitStatusList{
					unitStatus("example-foo@1.service", "@1", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "loaded", "loaded"),
					unitStatus("example-bar@1.service", "@1", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "loaded", "loaded"),
					unitStatus("example-foo@2.service", "@2", "172.17.8.102", "9ebb53b04b0d46fb94b4fd1b3f562d2b", "loaded", "loaded"),
					unitStatus("example-bar@2.service", "@2", "172.17.8.102", "9ebb53b04b0d46fb94b4fd1b3f562d2b", "loaded", "loaded"),
					unitStatus("example-foo@3.service", "@3", "172.17.8.103", "e3cb5f13a9164ba5b7eff6c920475e61", "loaded", "loaded"),
					unitStatus("example-bar@3.service", "@3", "172.17.8.103", "e3cb5f13a9164ba5b7eff6c920475e61", "loaded", "loaded"),
				},
				Verbose: true,
			},
		},
		// 1 slice of 2 groups with 2 units - only show example group
		testCase{
			Comment: "1 slice of 2 groups with 2 units - only show example group",
			Expected: []string{
				"Group | Units | FDState | FCState | SAState | IP | Machine",
				"",
				"example@1 | * | loaded | loaded | inactive | 172.17.8.101 | 505e0d7802d7439a924c269b76f34b5f",
				"",
			},
			Input: input{
				Group: "example",
				USL: controller.UnitStatusList{
					unitStatus("example-foo@1.service", "@1", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "loaded", "loaded"),
					unitStatus("example-bar@1.service", "@1", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "loaded", "loaded"),
					unitStatus("myapp-foo@1.service", "@1", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "loaded", "loaded"),
					unitStatus("myapp-bar@1.service", "@1", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "loaded", "loaded"),
				},
				Verbose: false,
			},
		},
		// One group contains different statuses => expand view
		testCase{
			Comment: "One group contains different statuses => expand view",
			Expected: []string{
				"Group | Units | FDState | FCState | SAState | IP | Machine",
				"",
				"example@1 | example-foo@1.service | launched | loaded | inactive | 172.17.8.101 | 505e0d7802d7439a924c269b76f34b5f",
				"example@1 | example-bar@1.service | launched | launched | inactive | 172.17.8.101 | 505e0d7802d7439a924c269b76f34b5f",
				"example@2 | * | launched | launched | inactive | 172.17.8.101 | 505e0d7802d7439a924c269b76f34b5f",
				"",
			},
			Input: input{
				Group: "example",
				USL: controller.UnitStatusList{
					unitStatus("example-foo@1.service", "@1", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "loaded", "launched"),
					unitStatus("example-bar@1.service", "@1", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "launched", "launched"),
					unitStatus("example-foo@2.service", "@2", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "launched", "launched"),
					unitStatus("example-bar@2.service", "@2", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "launched", "launched"),
				},
				Verbose: false,
			},
		},
	}

	// execute test cases
	for _, test := range testCases {
		globalFlags.Verbose = test.Input.Verbose

		got, err := createStatus(test.Input.Group, test.Input.USL)
		Expect(err).To(Not(HaveOccurred()))

		Expect(got).To(Equal(test.Expected), test.Comment)
	}
}

func unitStatus(name, slice, machineIP, machineID, currentState, desiredState string) fleet.UnitStatus {
	return fleet.UnitStatus{
		Current: currentState,
		Desired: desiredState,
		Machine: []fleet.MachineStatus{
			fleet.MachineStatus{
				ID:            machineID,
				IP:            net.ParseIP(machineIP),
				SystemdActive: "inactive",
				SystemdSub:    "inactive",
				UnitHash:      "4311",
			},
		},
		Name:  name,
		Slice: slice,
	}
}
