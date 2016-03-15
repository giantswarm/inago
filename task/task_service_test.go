package task

import (
	"fmt"
	"testing"
	"time"

	"golang.org/x/net/context"
)

func Test_Task_TaskService_Create_Success(t *testing.T) {
	newTaskService := NewTaskService(DefaultConfig())

	testData := "invalid"

	action := func(ctx context.Context) error {
		testData = "valid"
		return nil
	}

	taskObject, err := newTaskService.Create(context.Background(), action)
	if err != nil {
		t.Fatalf("TaskService.Create did return error: %#v", err)
	}

	taskObject, err = newTaskService.WaitForFinalStatus(context.Background(), taskObject.ID, nil)
	if err != nil {
		t.Fatalf("TaskService.WaitForFinalStatus did return error: %#v", err)
	}

	if !HasFinalStatus(taskObject) {
		t.Fatalf("received task object did NOT have a final status")
	}

	if testData != "valid" {
		t.Fatalf("test data did NOT have a valid value")
	}

	if taskObject.Error != "" {
		t.Fatalf("received task object did have a error")
	}
}

func Test_Task_TastService_Create_Error(t *testing.T) {
	newConfig := DefaultConfig()
	newConfig.WaitSleep = 10 * time.Millisecond
	newTaskService := NewTaskService(newConfig)

	action := func(ctx context.Context) error {
		return fmt.Errorf("test error")
	}

	taskObject, err := newTaskService.Create(context.Background(), action)
	if err != nil {
		t.Fatalf("TaskService.Create did return error: %#v", err)
	}

	taskObject, err = newTaskService.WaitForFinalStatus(context.Background(), taskObject.ID, nil)
	if err != nil {
		t.Fatalf("TaskService.WaitForFinalStatus did return error: %#v", err)
	}

	if !HasFinalStatus(taskObject) {
		t.Fatalf("received task object did NOT have a final status")
	}

	if taskObject.Error != "test error" {
		t.Fatalf("received task object did NOT have a proper error")
	}
}

func Test_Task_TastService_Create_FetchState(t *testing.T) {
	newConfig := DefaultConfig()
	newConfig.WaitSleep = 10 * time.Millisecond
	newTaskService := NewTaskService(newConfig)

	action := func(ctx context.Context) error {
		return nil
	}

	taskObject, err := newTaskService.Create(context.Background(), action)
	if err != nil {
		t.Fatalf("TaskService.Create did return error: %#v", err)
	}

	taskObject, err = newTaskService.WaitForFinalStatus(context.Background(), taskObject.ID, nil)
	if err != nil {
		t.Fatalf("TaskService.WaitForFinalStatus did return error: %#v", err)
	}

	if !HasFinalStatus(taskObject) {
		t.Fatalf("received task object did NOT have a final status")
	}

	// Fetching invalid state should not work.
	_, err = newTaskService.FetchState(context.Background(), "invalid")
	if !IsTaskObjectNotFound(err) {
		t.Fatalf("TaskService.Create did NOT return proper error")
	}

	// Fetching valid state should work.
	taskObject, err = newTaskService.FetchState(context.Background(), taskObject.ID)
	if err != nil {
		t.Fatalf("TaskService.Create did return error: %#v", err)
	}

	if !HasFinalStatus(taskObject) {
		t.Fatalf("received task object did NOT have a final status")
	}
}

func Test_Task_TastService_Create_Wait(t *testing.T) {
	newConfig := DefaultConfig()
	newConfig.WaitSleep = 10 * time.Millisecond
	newTaskService := NewTaskService(newConfig)

	action := func(ctx context.Context) error {
		// Just something to do, so the task blocks
		time.Sleep(300 * time.Millisecond)

		return nil
	}

	originalTaskObject, err := newTaskService.Create(context.Background(), action)
	if err != nil {
		t.Fatalf("TaskService.Create did return error: %#v", err)
	}

	// Directly close and end waiting.
	closer := make(chan struct{}, 1)
	closer <- struct{}{}

	taskObject, err := newTaskService.WaitForFinalStatus(context.Background(), originalTaskObject.ID, closer)
	if err != nil {
		t.Fatalf("TaskService.WaitForFinalStatus did return error: %#v", err)
	}
	if taskObject != nil {
		t.Fatalf("Expected canceled WaitForFinalStatus to return nil, nil")
	}

	taskObject, err = newTaskService.FetchState(context.Background(), originalTaskObject.ID)
	if err != nil {
		t.Fatalf("TaskService.FetchState did return error: %#v", err)
	}

	// When we use the closer to end waiting before the task is finished, the
	// task object should not have a final state yet.
	if HasFinalStatus(taskObject) {
		t.Fatalf("received task object did have a final status")
	}
}
