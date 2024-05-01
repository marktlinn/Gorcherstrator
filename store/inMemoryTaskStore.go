package store

import (
	"fmt"

	"github.com/marktlinn/Gorcherstrator/task"
)

// InMemoryTaskStore is a map structure store of Tasks held in memory.
type InMemoryTaskStore struct {
	DB map[string]*task.Task
}

// NewInMemoryTaskStore creates a new InMemoryTaskStore and returned a reference to it.
func NewInMemoryTaskStore() *InMemoryTaskStore {
	return &InMemoryTaskStore{
		DB: make(map[string]*task.Task),
	}
}

func (i *InMemoryTaskStore) Get(key string) (any, error) {
	// TODO: complete.
	return nil, nil
}

// Put inserts a key value pair into the NewInMemoryTaskStore, asserting first that the value is a pointer to a task.Task.
func (i *InMemoryTaskStore) Put(key string, value any) error {
	task, ok := value.(*task.Task)
	if !ok {
		return fmt.Errorf("failed to assert value %s as type *task.Task\n", value)
	}
	i.DB[key] = task
	return nil
}
