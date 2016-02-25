package task

type Backend interface {
	Get(taskID string) (*TaskObject, error)
	Set(taskObject *TaskObject) error
}
