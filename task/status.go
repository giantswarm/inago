package task

// ActiveStatus represents a status indicating activation or deactivation.
type ActiveStatus string

const (
	StatusStarted ActiveStatus = "started"
	StatusStopped ActiveStatus = "stopped"
)

// FinalStatus represents any status that is final. A task having this status
// will not change its status anymore.
type FinalStatus string

const (
	StatusFailed    FinalStatus = "failed"
	StatusSucceeded FinalStatus = "succeeded"
)

// HasFailedStatus determines whether a task has failed or not. Note that this
// is about a final status.
func HasFailedStatus(taskObject *TaskObject) bool {
	if taskObject.ActiveStatus == StatusStopped && taskObject.FinalStatus == StatusFailed {
		return true
	}

	return false
}

// HasFinalStatus determines whether a task has a final status or not.
func HasFinalStatus(taskObject *TaskObject) bool {
	if HasFailedStatus(taskObject) || HasSucceededStatus(taskObject) {
		return true
	}

	return false
}

// HasSucceededStatus determines whether a task has succeeded or not. Note that
// this is about a final status.
func HasSucceededStatus(taskObject *TaskObject) bool {
	if taskObject.ActiveStatus == StatusStopped && taskObject.FinalStatus == StatusSucceeded {
		return true
	}

	return false
}
