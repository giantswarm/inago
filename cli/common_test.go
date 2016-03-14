package cli

import (
	"net"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/giantswarm/inago/controller"
	"github.com/giantswarm/inago/fleet"
)

func Test_Common_createStatus(t *testing.T) {
	RegisterTestingT(t)

	type input struct {
		Group   string
		USL     controller.UnitStatusList
		Verbose bool
	}
	type testCase struct {
		Comment  string
		Input    input
		Expected []string
	}

	testCases := []testCase{
		// 3 slices of 1 group with 2 units
		testCase{
			Comment: "3 slices of 1 group with 2 units",
			Input: input{
				Group: "example",
				USL: controller.UnitStatusList{
					loadedUnitStatus("example-foo@1.service", "1", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "loaded", "loaded"),
					loadedUnitStatus("example-bar@1.service", "1", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "loaded", "loaded"),
					loadedUnitStatus("example-foo@2.service", "2", "172.17.8.102", "9ebb53b04b0d46fb94b4fd1b3f562d2b", "loaded", "loaded"),
					loadedUnitStatus("example-bar@2.service", "2", "172.17.8.102", "9ebb53b04b0d46fb94b4fd1b3f562d2b", "loaded", "loaded"),
					loadedUnitStatus("example-foo@3.service", "3", "172.17.8.103", "e3cb5f13a9164ba5b7eff6c920475e61", "loaded", "loaded"),
					loadedUnitStatus("example-bar@3.service", "3", "172.17.8.103", "e3cb5f13a9164ba5b7eff6c920475e61", "loaded", "loaded"),
				},
				Verbose: false,
			},
			Expected: []string{
				"Group | Units | FDState | FCState | SAState | IP | Machine",
				"",
				"example@1 | * | loaded | loaded | inactive | 172.17.8.101 | 505e0d7802d7439a924c269b76f34b5f",
				"example@2 | * | loaded | loaded | inactive | 172.17.8.102 | 9ebb53b04b0d46fb94b4fd1b3f562d2b",
				"example@3 | * | loaded | loaded | inactive | 172.17.8.103 | e3cb5f13a9164ba5b7eff6c920475e61",
				"",
			},
		},
		// 1 slice of 1 group with 2 units
		testCase{
			Comment: "1 slice of 1 group with 2 units",
			Input: input{
				Group: "example",
				USL: controller.UnitStatusList{
					loadedUnitStatus("example-foo", "1", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "loaded", "loaded"),
					loadedUnitStatus("example-bar", "1", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "loaded", "loaded"),
				},
				Verbose: false,
			},
			Expected: []string{
				"Group | Units | FDState | FCState | SAState | IP | Machine",
				"",
				"example@1 | * | loaded | loaded | inactive | 172.17.8.101 | 505e0d7802d7439a924c269b76f34b5f",
				"",
			},
		},
		// 3 slices of 1 group with 2 units - verbose
		testCase{
			Comment: "3 slices of 1 group with 2 units - verbose",
			Input: input{
				Group: "example",
				USL: controller.UnitStatusList{
					loadedUnitStatus("example-foo@1.service", "1", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "loaded", "loaded"),
					loadedUnitStatus("example-bar@1.service", "1", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "loaded", "loaded"),
					loadedUnitStatus("example-foo@2.service", "2", "172.17.8.102", "9ebb53b04b0d46fb94b4fd1b3f562d2b", "loaded", "loaded"),
					loadedUnitStatus("example-bar@2.service", "2", "172.17.8.102", "9ebb53b04b0d46fb94b4fd1b3f562d2b", "loaded", "loaded"),
					loadedUnitStatus("example-foo@3.service", "3", "172.17.8.103", "e3cb5f13a9164ba5b7eff6c920475e61", "loaded", "loaded"),
					loadedUnitStatus("example-bar@3.service", "3", "172.17.8.103", "e3cb5f13a9164ba5b7eff6c920475e61", "loaded", "loaded"),
				},
				Verbose: true,
			},
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
		},
		// 1 slice of 2 groups with 2 units - only show example group
		testCase{
			Comment: "1 slice of 2 groups with 2 units - only show example group",
			Input: input{
				Group: "example",
				USL: controller.UnitStatusList{
					loadedUnitStatus("example-foo@1.service", "1", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "loaded", "loaded"),
					loadedUnitStatus("example-bar@1.service", "1", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "loaded", "loaded"),
					loadedUnitStatus("myapp-foo@1.service", "1", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "loaded", "loaded"),
					loadedUnitStatus("myapp-bar@1.service", "1", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "loaded", "loaded"),
				},
				Verbose: false,
			},
			Expected: []string{
				"Group | Units | FDState | FCState | SAState | IP | Machine",
				"",
				"example@1 | * | loaded | loaded | inactive | 172.17.8.101 | 505e0d7802d7439a924c269b76f34b5f",
				"",
			},
		},
		// One group contains different statuses => expand view
		testCase{
			Comment: "One group contains different statuses => expand view",
			Input: input{
				Group: "example",
				USL: controller.UnitStatusList{
					loadedUnitStatus("example-foo@1.service", "1", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "loaded", "launched"),
					loadedUnitStatus("example-bar@1.service", "1", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "launched", "launched"),
					loadedUnitStatus("example-foo@2.service", "2", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "launched", "launched"),
					loadedUnitStatus("example-bar@2.service", "2", "172.17.8.101", "505e0d7802d7439a924c269b76f34b5f", "launched", "launched"),
				},
				Verbose: false,
			},
			Expected: []string{
				"Group | Units | FDState | FCState | SAState | IP | Machine",
				"",
				"example@1 | example-foo@1.service | launched | loaded | inactive | 172.17.8.101 | 505e0d7802d7439a924c269b76f34b5f",
				"example@1 | example-bar@1.service | launched | launched | inactive | 172.17.8.101 | 505e0d7802d7439a924c269b76f34b5f",
				"example@2 | * | launched | launched | inactive | 172.17.8.101 | 505e0d7802d7439a924c269b76f34b5f",
				"",
			},
		},

		// A group who is NOT loaded onto a machine
		testCase{
			Comment: "A group with a MachineState should still render (verbose)",
			Input: input{
				Group: "example",
				USL: controller.UnitStatusList{
					unloadedUnitStatus("example-1@1.service", "1", "active"),
					unloadedUnitStatus("example-2@1.service", "1", "active"),
				},
				Verbose: true,
			},
			Expected: []string{
				"Group | Units | FDState | FCState | SAState | Hash | IP | Machine",
				"",
				"example@1 | example-1@1.service | active | inactive | - | - | - | -",
				"example@1 | example-2@1.service | active | inactive | - | - | - | -",
				"",
			},
		},

		// A group who is NOT loaded onto a machine
		testCase{
			Comment: "A group with a MachineState should still render",
			Input: input{
				Group: "example",
				USL: controller.UnitStatusList{
					unloadedUnitStatus("example-1@1.service", "1", "active"),
					unloadedUnitStatus("example-2@1.service", "1", "active"),
				},
				Verbose: false,
			},
			Expected: []string{
				"Group | Units | FDState | FCState | SAState | IP | Machine",
				"",
				"example@1 | * | active | inactive | - | - | -",
				"",
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

func loadedUnitStatus(name, sliceID, machineIP, machineID, currentState, desiredState string) fleet.UnitStatus {
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
		Name:    name,
		SliceID: sliceID,
	}
}

func unloadedUnitStatus(name, sliceID, desiredState string) fleet.UnitStatus {
	return fleet.UnitStatus{
		Current: "inactive",
		Desired: desiredState,
		Machine: []fleet.MachineStatus{},
		Name:    name,
		SliceID: sliceID,
	}
}
