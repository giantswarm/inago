package task

import (
	"testing"
)

func Test_Task_Status_HasFinalStatus(t *testing.T) {
	testCases := []struct {
		Input    *TaskObject
		Expected bool
	}{
		// This status combination is invalid.
		{
			Input: &TaskObject{
				ActiveStatus: StatusStarted,
				Error:        "",
				FinalStatus:  StatusFailed,
				ID:           "",
			},
			Expected: false,
		},

		{
			Input: &TaskObject{
				ActiveStatus: StatusStopped,
				Error:        "",
				FinalStatus:  StatusFailed,
				ID:           "",
			},
			Expected: true,
		},

		// This status combination is invalid.
		{
			Input: &TaskObject{
				ActiveStatus: StatusStarted,
				Error:        "",
				FinalStatus:  StatusSucceeded,
				ID:           "",
			},
			Expected: false,
		},
		{
			Input: &TaskObject{
				ActiveStatus: StatusStopped,
				Error:        "",
				FinalStatus:  StatusSucceeded,
				ID:           "",
			},
			Expected: true,
		},
		{
			Input: &TaskObject{
				ActiveStatus: StatusStarted,
				Error:        "",
				FinalStatus:  "",
				ID:           "",
			},
			Expected: false,
		},
		{
			Input: &TaskObject{
				ActiveStatus: StatusStopped,
				Error:        "",
				FinalStatus:  "",
				ID:           "",
			},
			Expected: false,
		},
		{
			Input: &TaskObject{
				ActiveStatus: StatusStarted,
				Error:        "",
				FinalStatus:  "",
				ID:           "",
			},
			Expected: false,
		},
		{
			Input: &TaskObject{
				ActiveStatus: StatusStopped,
				Error:        "",
				FinalStatus:  "",
				ID:           "",
			},
			Expected: false,
		},

		{
			Input: &TaskObject{
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
