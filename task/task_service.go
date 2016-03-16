package task

import (
	"time"

	"github.com/satori/go.uuid"
	"golang.org/x/net/context"

	"github.com/giantswarm/inago/logging"
)

const (
	// ContextTaskID is the key for the current task-id stored in the context.Context when executing tasks.
	ContextTaskID = "task-id"
)

// Action represents any work to be done when executing a task.
type Action func(ctx context.Context) error

// Task represents a task that is executable.
type Task struct {
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

// Service represents a task managing unit being able to act on task
// objects.
type Service interface {
	// Create creates a new task object configured with the given action. The
	// task object is immediately returned and its corresponding action is
	// executed asynchronously.
	Create(ctx context.Context, action Action) (*Task, error)

	// FetchState fetches and returns the current state and status for the given
	// task ID.
	FetchState(ctx context.Context, taskID string) (*Task, error)

	// MarkAsSucceeded marks the task object as succeeded and persists its state.
	// The returned task object is actually the refreshed version of the provided
	// one.
	MarkAsSucceeded(ctx context.Context, taskObject *Task) (*Task, error)

	// MarkAsFailedWithError marks the task object as failed, adds information of
	// thegiven error and persists the task objects's state. The returned task
	// object is actually the refreshed version of the provided one.
	MarkAsFailedWithError(ctx context.Context, taskObject *Task, err error) (*Task, error)

	// PersistState writes the given task object to the configured Storage.
	PersistState(ctx context.Context, taskObject *Task) error

	// WaitForFinalStatus blocks and waits for the given task to reach a final
	// status. The given closer can end the waiting and thus stop blocking the
	// call to WaitForFinalStatus.
	WaitForFinalStatus(ctx context.Context, taskID string, closer <-chan struct{}) (*Task, error)
}

// Config represents the configurations for the task service that is
// going to be created.
type Config struct {
	Storage Storage

	// WaitSleep represents the time to sleep between state-check cycles.
	WaitSleep time.Duration

	// Logger provides an initialised logger.
	Logger logging.Logger
}

// DefaultConfig returns a best effort default configuration for the
// task service.
func DefaultConfig() Config {
	newConfig := Config{
		Storage:   NewMemoryStorage(),
		WaitSleep: 1 * time.Second,
		Logger:    logging.NewLogger(logging.DefaultConfig()),
	}

	return newConfig
}

// NewTaskService returns a new configured task service instance.
func NewTaskService(config Config) Service {
	newTaskService := &taskService{
		Config: config,
	}

	return newTaskService
}

type taskService struct {
	Config
}

func (ts *taskService) Create(ctx context.Context, action Action) (*Task, error) {
	taskID := uuid.NewV4().String()
	ctx = context.WithValue(ctx, ContextTaskID, taskID)
	ts.Config.Logger.Debug(ctx, "task: creating task")

	taskObject := &Task{
		ID:           taskID,
		ActiveStatus: StatusStarted,
		FinalStatus:  "",
	}

	go func(ctx context.Context) {
		err := action(ctx)
		if err != nil {
			_, markErr := ts.MarkAsFailedWithError(ctx, taskObject, err)
			if markErr != nil {
				ts.Config.Logger.Error(nil, "[E] Task.MarkAsFailed failed: %#v", maskAny(markErr))
				return
			}
			return
		}

		_, err = ts.MarkAsSucceeded(ctx, taskObject)
		if err != nil {
			ts.Config.Logger.Error(nil, "[E] Task.MarkAsSucceeded failed: %#v", maskAny(err))
			return
		}
	}(ctx)

	err := ts.PersistState(ctx, taskObject)
	if err != nil {
		return nil, maskAny(err)
	}

	ts.Config.Logger.Debug(ctx, "task: created task: %v", taskObject.ID)

	return taskObject, nil
}

func (ts *taskService) FetchState(ctx context.Context, taskID string) (*Task, error) {
	ts.Config.Logger.Debug(ctx, "task: fetching state for task: %v", taskID)

	var err error

	taskObject, err := ts.Storage.Get(taskID)
	if err != nil {
		return nil, maskAny(err)
	}

	return taskObject, nil
}

func (ts *taskService) MarkAsFailedWithError(ctx context.Context, taskObject *Task, err error) (*Task, error) {
	ts.Config.Logger.Debug(ctx, "task: marking as failed for task: %v", taskObject.ID)

	taskObject.ActiveStatus = StatusStopped
	taskObject.Error = err.Error()
	taskObject.FinalStatus = StatusFailed

	err = ts.PersistState(ctx, taskObject)
	if err != nil {
		return nil, maskAny(err)
	}

	return taskObject, nil
}

func (ts *taskService) MarkAsSucceeded(ctx context.Context, taskObject *Task) (*Task, error) {
	ts.Config.Logger.Debug(ctx, "task: marking as succeeded for task: %v", taskObject.ID)

	taskObject.ActiveStatus = StatusStopped
	taskObject.FinalStatus = StatusSucceeded

	err := ts.PersistState(ctx, taskObject)
	if err != nil {
		return nil, maskAny(err)
	}

	return taskObject, nil
}

func (ts *taskService) PersistState(ctx context.Context, taskObject *Task) error {
	ts.Config.Logger.Debug(ctx, "task: persisting state for task: %v", taskObject.ID)

	err := ts.Storage.Set(taskObject)
	if err != nil {
		return maskAny(err)
	}

	return nil
}

// WaitForFinalStatus acts as described in the interface comments. Note that
// both, task object and error will be nil in case the closer ends waiting for
// the task to reach a final state.
func (ts *taskService) WaitForFinalStatus(ctx context.Context, taskID string, closer <-chan struct{}) (*Task, error) {
	ts.Config.Logger.Debug(ctx, "task: waiting for final status for task: %v", taskID)

	for {
		select {
		case <-closer:
			return nil, nil
		case <-time.After(ts.WaitSleep):
			taskObject, err := ts.FetchState(ctx, taskID)
			if err != nil {
				return nil, maskAny(err)
			}

			if HasFinalStatus(taskObject) {
				return taskObject, nil
			}
		}
	}
}
