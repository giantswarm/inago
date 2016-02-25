package task

import (
	"fmt"
	"time"

	"github.com/satori/go.uuid"
)

// Action represents any work to be done when executing a task.
type Action func() error

// TaskObject represents a task that is executable.
type TaskObject struct {
	// ActiveStatus represents a status indicating activation or deactivation.
	ActiveStatus ActiveStatus

	// Error represents the message of an error occured during task execution, if
	// any.
	Error string

	// FinalStatus represents any status that is final. A task having this status
	// will not change its status anymore.
	FinalStatus FinalStatus

	// ID represents the task identifier.
	ID string
}

// TaskService represents a task managing unit being able to act on task
// objects.
type TaskService interface {
	// Create creates a new task object configured with the given action. The
	// task object is immediately returned and its corresponding action is
	// executed asynchronously.
	Create(action Action) (*TaskObject, error)

	// FetchState fetches and returns the current state and status for the given
	// task ID.
	FetchState(taskID string) (*TaskObject, error)

	// MarkAsSucceeded marks the task object as succeeded and persists its state.
	// The returned task object is actually the refreshed version of the provided
	// one.
	MarkAsSucceeded(taskObject *TaskObject) (*TaskObject, error)

	// MarkAsFailedWithError marks the task object as failed, adds information of
	// thegiven error and persists the task objects's state. The returned task
	// object is actually the refreshed version of the provided one.
	MarkAsFailedWithError(taskObject *TaskObject, err error) (*TaskObject, error)

	// PersistState writes the given task object to the configured Storage.
	PersistState(taskObject *TaskObject) error

	// WaitForFinalStatus blocks and waits for the given task to reach a final
	// status. The given closer can end the waiting and thus stop blocking the
	// call to WaitForFinalStatus.
	WaitForFinalStatus(taskID string, closer <-chan struct{}) (*TaskObject, error)
}

// TaskServiceConfig represents the configurations for the task service that is
// going to be created.
type TaskServiceConfig struct {
	Storage Storage

	// WaitSleep represents the time to sleep between state-check cycles.
	WaitSleep time.Duration
}

// DefaultTaskServiceConfig returns a best effort default configuration for the
// task service.
func DefaultTaskServiceConfig() TaskServiceConfig {
	newConfig := TaskServiceConfig{
		Storage:   NewMemoryStorage(),
		WaitSleep: 1 * time.Second,
	}

	return newConfig
}

// NewTaskService returns a new configured task service instance.
func NewTaskService(config TaskServiceConfig) TaskService {
	newTaskService := &taskService{
		TaskServiceConfig: config,
	}

	return newTaskService
}

type taskService struct {
	TaskServiceConfig
}

func (ts *taskService) Create(action Action) (*TaskObject, error) {
	taskObject := &TaskObject{
		ID:           uuid.NewV4().String(),
		ActiveStatus: StatusStarted,
		FinalStatus:  "",
	}

	go func() {
		err := action()
		if err != nil {
			_, markErr := ts.MarkAsFailedWithError(taskObject, err)
			if markErr != nil {
				fmt.Printf("[E] Task.MarkAsFailed failed: %#v\n", maskAny(markErr))
				return
			}
			return
		}

		_, err = ts.MarkAsSucceeded(taskObject)
		if err != nil {
			fmt.Printf("[E] Task.MarkAsSucceeded failed: %#v\n", maskAny(err))
			return
		}
	}()

	err := ts.PersistState(taskObject)
	if err != nil {
		return nil, maskAny(err)
	}

	return taskObject, nil
}

func (ts *taskService) FetchState(taskID string) (*TaskObject, error) {
	var err error

	taskObject, err := ts.Storage.Get(taskID)
	if err != nil {
		return nil, maskAny(err)
	}

	return taskObject, nil
}

func (ts *taskService) MarkAsFailedWithError(taskObject *TaskObject, err error) (*TaskObject, error) {
	taskObject.ActiveStatus = StatusStopped
	taskObject.Error = err.Error()
	taskObject.FinalStatus = StatusFailed

	err = ts.PersistState(taskObject)
	if err != nil {
		return nil, maskAny(err)
	}

	return taskObject, nil
}

func (ts *taskService) MarkAsSucceeded(taskObject *TaskObject) (*TaskObject, error) {
	taskObject.ActiveStatus = StatusStopped
	taskObject.FinalStatus = StatusSucceeded

	err := ts.PersistState(taskObject)
	if err != nil {
		return nil, maskAny(err)
	}

	return taskObject, nil
}

func (ts *taskService) PersistState(taskObject *TaskObject) error {
	err := ts.Storage.Set(taskObject)
	if err != nil {
		return maskAny(err)
	}

	return nil
}

// WaitForFinalStatus acts as described in the interface comments. Note that
// both, task object and error will be nil in case the closer ends waiting for
// the task to reach a final state.
func (ts *taskService) WaitForFinalStatus(taskID string, closer <-chan struct{}) (*TaskObject, error) {
	for {
		select {
		case <-closer:
			return nil, nil
		default:
			taskObject, err := ts.FetchState(taskID)
			if err != nil {
				return nil, maskAny(err)
			}

			if HasFinalStatus(taskObject) {
				return taskObject, nil
			}
		}

		time.Sleep(ts.WaitSleep)
	}
}
