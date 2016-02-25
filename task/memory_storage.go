package task

import (
	"sync"
)

// NewMemoryStorage creates a backend implementation for pseudo in-memory
// persistence.
func NewMemoryStorage() Storage {
	newStorage := &memoryStorage{
		Mutex:   sync.Mutex{},
		Storage: map[string]*TaskObject{},
	}

	return newStorage
}

type memoryStorage struct {
	Mutex   sync.Mutex
	Storage map[string]*TaskObject
}

func (mb *memoryStorage) Get(taskID string) (*TaskObject, error) {
	mb.Mutex.Lock()
	defer mb.Mutex.Unlock()

	if to, ok := mb.Storage[taskID]; ok {
		return to, nil
	}

	return nil, maskAny(taskObjectNotFoundError)
}

func (mb *memoryStorage) Set(taskObject *TaskObject) error {
	mb.Mutex.Lock()
	defer mb.Mutex.Unlock()

	mb.Storage[taskObject.ID] = taskObject

	return nil
}
