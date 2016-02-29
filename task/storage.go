package task

// Storage represents some storage solution to persist task objects.
type Storage interface {
	// Get fetches the corresponding task object for the given task ID.
	Get(taskID string) (*Task, error)

	// Set persists the given task object for its corresponding task ID.
	Set(taskObject *Task) error
}
