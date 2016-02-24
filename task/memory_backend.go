package task

import (
	"sync"
)

func NewMemoryBackend() Backend {
	newBackend := &memoryBackend{
		Mutex:   sync.Mutex{},
		Storage: map[string]*TaskObject{},
	}

	return newBackend
}

type memoryBackend struct {
	Mutex   sync.Mutex
	Storage map[string]*TaskObject
}

func (mb *memoryBackend) Get(taskID string) (*TaskObject, error) {
	mb.Mutex.Lock()
	defer mb.Mutex.Unlock()

	if to, ok := mb.Storage[taskID]; ok {
		return to, nil
	}

	return nil, maskAny(taskObjectNotFoundError)
}

func (mb *memoryBackend) Set(taskObject *TaskObject) error {
	mb.Mutex.Lock()
	defer mb.Mutex.Unlock()

	mb.Storage[taskObject.ID] = taskObject

	return nil
}
