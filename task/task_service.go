package task

import (
	"fmt"
	"time"

	"github.com/satori/go.uuid"
)

type Action func() error

type TaskObject struct {
	ID           string
	ActiveStatus ActiveStatus
	Error        string
	FinalStatus  FinalStatus
}

type TaskService interface {
	Create(action Action) (*TaskObject, error)
	FetchState(taskObject *TaskObject) (*TaskObject, error)
	MarkAsSucceeded(taskObject *TaskObject) (*TaskObject, error)
	MarkAsFailedWithError(taskObject *TaskObject, err error) (*TaskObject, error)
	PersistState(taskObject *TaskObject) error
	WaitForFinalStatus(taskObject *TaskObject, closer <-chan struct{}) (*TaskObject, error)
}

type TaskServiceConfig struct {
	Backend Backend
}

func DefaultTaskServiceConfig() TaskServiceConfig {
	newConfig := TaskServiceConfig{
		Backend: NewMemoryBackend(),
	}

	return newConfig
}

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

func (ts *taskService) FetchState(taskObject *TaskObject) (*TaskObject, error) {
	var err error

	taskObject, err = ts.Backend.Get(taskObject.ID)
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
	err := ts.Backend.Set(taskObject)
	if err != nil {
		return maskAny(err)
	}

	return nil
}

func (ts *taskService) WaitForFinalStatus(taskObject *TaskObject, closer <-chan struct{}) (*TaskObject, error) {
	for {
		select {
		case <-closer:
			return taskObject, nil
		default:
			taskObject, err := ts.FetchState(taskObject)
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
