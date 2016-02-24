package task

import (
	"fmt"
	"time"

	"github.com/satori/go.uuid"
)

type Action func() error

type TaskObject struct {
	ID string

	ActiveStatus ActiveStatus
	FinalStatus  FinalStatus
}

type Task interface {
	Create(action Action) (*TaskObject, error)
	FetchState(taskObject *TaskObject) (*TaskObject, error)
	MarkAsSucceeded(taskObject *TaskObject) (*TaskObject, error)
	MarkAsFailed(taskObject *TaskObject) (*TaskObject, error)
	PersistState(taskObject *TaskObject) error
	WaitForFinalStatus(taskObject *TaskObject, closer <-chan struct{}) (*TaskObject, error)
}

type TaskConfig struct {
	Backend Backend
}

func DefaultTaskConfig() TaskConfig {
	newConfig := TaskConfig{
		Backend: NewMemoryBackend(),
	}

	return newConfig
}

func NewTask(config TaskConfig) Task {
	newTask := &task{
		TaskConfig: config,
	}

	return newTask
}

type task struct {
	TaskConfig
}

func (t *task) Create(action Action) (*TaskObject, error) {
	taskObject := &TaskObject{
		ID:           uuid.NewV4().String(),
		ActiveStatus: StatusStarted,
		FinalStatus:  "",
	}

	go func() {
		err := action()
		if err != nil {
			fmt.Printf("[E] Task.Action failed: %#v\n", maskAny(err))
			return
		}

		_, err = t.MarkAsFailed(taskObject)
		if err != nil {
			fmt.Printf("[E] Task.MarkAsFailed failed: %#v\n", maskAny(err))
			return
		}

		_, err = t.MarkAsSucceeded(taskObject)
		if err != nil {
			fmt.Printf("[E] Task.MarkAsSucceeded failed: %#v\n", maskAny(err))
			return
		}
	}()

	err := t.PersistState(taskObject)
	if err != nil {
		return nil, maskAny(err)
	}

	return taskObject, nil
}

func (t *task) FetchState(taskObject *TaskObject) (*TaskObject, error) {
	var err error

	taskObject, err = t.Backend.Get(taskObject.ID)
	if err != nil {
		return nil, maskAny(err)
	}

	return taskObject, nil
}

func (t *task) MarkAsFailed(taskObject *TaskObject) (*TaskObject, error) {
	taskObject.ActiveStatus = StatusStopped
	taskObject.FinalStatus = StatusFailed

	err := t.PersistState(taskObject)
	if err != nil {
		return nil, maskAny(err)
	}

	return taskObject, nil
}

func (t *task) MarkAsSucceeded(taskObject *TaskObject) (*TaskObject, error) {
	taskObject.ActiveStatus = StatusStopped
	taskObject.FinalStatus = StatusSucceeded

	err := t.PersistState(taskObject)
	if err != nil {
		return nil, maskAny(err)
	}

	return taskObject, nil
}

func (t *task) PersistState(taskObject *TaskObject) error {
	err := t.Backend.Set(taskObject)
	if err != nil {
		return maskAny(err)
	}

	return nil
}

func (t *task) WaitForFinalStatus(taskObject *TaskObject, closer <-chan struct{}) (*TaskObject, error) {
	for {
		select {
		case <-closer:
			return taskObject, nil
		default:
			taskObject, err := t.FetchState(taskObject)
			if err != nil {
				return nil, maskAny(err)
			}

			if HasFinalStatus(taskObject) {
				return taskObject, nil
			}
		}

		time.Sleep(1 * time.Second)
	}

	return taskObject, nil
}
