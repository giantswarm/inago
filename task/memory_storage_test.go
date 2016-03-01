package task

import (
	"testing"
)

func Test_Task_Storage_Memory(t *testing.T) {
	newStorage := NewMemoryStorage()

	taskID := "task-id"

	_, err := newStorage.Get(taskID)
	if !IsTaskObjectNotFound(err) {
		t.Fatalf("Storage.Get did NOT return proper error")
	}

	taskObject := &Task{
		ID: taskID,
	}

	err = newStorage.Set(taskObject)
	if err != nil {
		t.Fatalf("Storage.Get did return error: %#v", err)
	}

	taskObject, err = newStorage.Get(taskID)
	if err != nil {
		t.Fatalf("Storage.Get did return error: %#v", err)
	}

	if taskObject.ID != taskID {
		t.Fatalf("received task object differs from original task object")
	}
}
