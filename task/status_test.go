package task

import (
	"testing"
)

func Test_Task_Status_HasFinalStatus(t *testing.T) {
	testCases := []struct {
		Input    *Task
		Expected bool
	}{
		// This status combination is invalid.
		{
			Input: &Task{
				ActiveStatus: StatusStarted,
				Error:        "",
				FinalStatus:  StatusFailed,
				ID:           "",
			},
			Expected: false,
		},

		{
			Input: &Task{
				ActiveStatus: StatusStopped,
				Error:        "",
				FinalStatus:  StatusFailed,
				ID:           "",
			},
			Expected: true,
		},

		// This status combination is invalid.
		{
			Input: &Task{
				ActiveStatus: StatusStarted,
				Error:        "",
				FinalStatus:  StatusSucceeded,
				ID:           "",
			},
			Expected: false,
		},
		{
			Input: &Task{
				ActiveStatus: StatusStopped,
				Error:        "",
				FinalStatus:  StatusSucceeded,
				ID:           "",
			},
			Expected: true,
		},
		{
			Input: &Task{
				ActiveStatus: StatusStarted,
				Error:        "",
				FinalStatus:  "",
				ID:           "",
			},
			Expected: false,
		},
		{
			Input: &Task{
				ActiveStatus: StatusStopped,
				Error:        "",
				FinalStatus:  "",
				ID:           "",
			},
			Expected: false,
		},
		{
			Input: &Task{
				ActiveStatus: StatusStarted,
				Error:        "",
				FinalStatus:  "",
				ID:           "",
			},
			Expected: false,
		},
		{
			Input: &Task{
				ActiveStatus: StatusStopped,
				Error:        "",
				FinalStatus:  "",
				ID:           "",
			},
			Expected: false,
		},

		{
			Input: &Task{
				ActiveStatus: StatusStopped,
				Error:        "test error",
				FinalStatus:  "",
				ID:           "test-id",
			},
			Expected: false,
		},
	}

	for i, testCase := range testCases {
		output := HasFinalStatus(testCase.Input)

		if output != testCase.Expected {
			t.Fatalf("test case %d: %t != %t", i+1, output, testCase.Expected)
		}
	}
}
