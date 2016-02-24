package task

type ActiveStatus string

const (
	StatusStarted ActiveStatus = "started"
	StatusStopped ActiveStatus = "stopped"
)

type FinalStatus string

const (
	StatusFailed    FinalStatus = "failed"
	StatusSucceeded FinalStatus = "succeeded"
)

func HasFailedStatus(taskObject *TaskObject) bool {
	if taskObject.ActiveStatus == StatusStopped && taskObject.FinalStatus == StatusFailed {
		return true
	}

	return false
}

func HasFinalStatus(taskObject *TaskObject) bool {
	if HasFailedStatus(taskObject) || HasSucceededStatus(taskObject) {
		return true
	}

	return false
}

func HasSucceededStatus(taskObject *TaskObject) bool {
	if taskObject.ActiveStatus == StatusStopped && taskObject.FinalStatus == StatusSucceeded {
		return true
	}

	return false
}
