package task

import (
	"fmt"
	"testing"
	"time"
)

func Test_Task_TastService_Create_Success(t *testing.T) {
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
	newTaskService := NewTaskService(DefaultTaskServiceConfig())

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
	newTaskService := NewTaskService(DefaultTaskServiceConfig())

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
	newTaskService := NewTaskService(DefaultTaskServiceConfig())

	action := func() error {
		fmt.Printf("%#v\n", 2)

		time.Sleep(1 * time.Second)
		fmt.Printf("%#v\n", 7)

		return nil
	}

	fmt.Printf("%#v\n", 1)

	taskObject, err := newTaskService.Create(action)
	if err != nil {
		t.Fatalf("TaskService.Create did return error: %#v", err)
	}

	closer := make(chan struct{})

	fmt.Printf("%#v\n", 3)

	go func() {
		time.Sleep(200 * time.Millisecond)
		closer <- struct{}{}
	}()

	fmt.Printf("%#v\n", 4)

	_, err = newTaskService.WaitForFinalStatus(taskObject.ID, closer)
	if err != nil {
		t.Fatalf("TaskService.WaitForFinalStatus did return error: %#v", err)
	}

	fmt.Printf("%#v\n", 5)

	taskObject, err = newTaskService.FetchState(taskObject.ID)
	if err != nil {
		t.Fatalf("TaskService.Create did return error: %#v", err)
	}

	fmt.Printf("%#v\n", 6)

	// When we use the closer to end waiting before the task is finished, the
	// task object should not have a final state yet.
	if HasFinalStatus(taskObject) {
		t.Fatalf("received task object did have a final status")
	}
}
