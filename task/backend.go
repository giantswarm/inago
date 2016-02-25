package task

// Backend represents some storage solution to persist task objects.
type Backend interface {
	// Get fetches the corresponding task object for the given task ID.
	Get(taskID string) (*TaskObject, error)

	// Set persists the given task object for its corresponding task ID.
	Set(taskObject *TaskObject) error
}
