package task

import (
	"fmt"
	"testing"
	"time"
)

func Test_Task_TaskService_Create_Success(t *testing.T) {
	newTaskService := NewTaskService(DefaultTaskServiceConfig())

	testData := "invalid"

	action := func() error {
		testData = "valid"
		return nil
	}

	taskObject, err := newTaskService.Create(action)
	if err != nil {
		t.Fatalf("TaskService.Create did return error: %#v", err)
	}

	taskObject, err = newTaskService.WaitForFinalStatus(taskObject.ID, nil)
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
	newConfig := DefaultTaskServiceConfig()
	newConfig.WaitSleep = 10 * time.Millisecond
	newTaskService := NewTaskService(newConfig)

	action := func() error {
		return fmt.Errorf("test error")
	}

	taskObject, err := newTaskService.Create(action)
	if err != nil {
		t.Fatalf("TaskService.Create did return error: %#v", err)
	}

	taskObject, err = newTaskService.WaitForFinalStatus(taskObject.ID, nil)
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
	newConfig := DefaultTaskServiceConfig()
	newConfig.WaitSleep = 10 * time.Millisecond
	newTaskService := NewTaskService(newConfig)

	action := func() error {
		return nil
	}

	taskObject, err := newTaskService.Create(action)
	if err != nil {
		t.Fatalf("TaskService.Create did return error: %#v", err)
	}

	taskObject, err = newTaskService.WaitForFinalStatus(taskObject.ID, nil)
	if err != nil {
		t.Fatalf("TaskService.WaitForFinalStatus did return error: %#v", err)
	}

	if !HasFinalStatus(taskObject) {
		t.Fatalf("received task object did NOT have a final status")
	}

	// Fetching invalid state should not work.
	_, err = newTaskService.FetchState("invalid")
	if !IsTaskObjectNotFound(err) {
		t.Fatalf("TaskService.Create did NOT return proper error")
	}

	// Fetching valid state should work.
	taskObject, err = newTaskService.FetchState(taskObject.ID)
	if err != nil {
		t.Fatalf("TaskService.Create did return error: %#v", err)
	}

	if !HasFinalStatus(taskObject) {
		t.Fatalf("received task object did NOT have a final status")
	}
}

func Test_Task_TastService_Create_Wait(t *testing.T) {
	newConfig := DefaultTaskServiceConfig()
	newConfig.WaitSleep = 10 * time.Millisecond
	newTaskService := NewTaskService(newConfig)

	action := func() error {
		// Just something to do, so the task blocks
		time.Sleep(300 * time.Millisecond)

		return nil
	}

	originalTaskObject, err := newTaskService.Create(action)
	if err != nil {
		t.Fatalf("TaskService.Create did return error: %#v", err)
	}

	closer := make(chan struct{})

	go func() {
		time.Sleep(100 * time.Millisecond)
		closer <- struct{}{}
	}()

	taskObject, err := newTaskService.WaitForFinalStatus(originalTaskObject.ID, closer)
	if err != nil {
		t.Fatalf("TaskService.WaitForFinalStatus did return error: %#v", err)
	}
	if taskObject != nil {
		t.Fatalf("Expected canceled WaitForFinalStatus to return nil, nil")
	}

	taskObject, err = newTaskService.FetchState(originalTaskObject.ID)
	if err != nil {
		t.Fatalf("TaskService.FetchState did return error: %#v", err)
	}

	// When we use the closer to end waiting before the task is finished, the
	// task object should not have a final state yet.
	if HasFinalStatus(taskObject) {
		t.Fatalf("received task object did have a final status")
	}
}
