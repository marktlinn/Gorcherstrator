package store

import (
	"fmt"

	"github.com/marktlinn/Gorcherstrator/task"
)

// InMemoryTaskStore is a map structure store of Tasks held in memory.
type InMemoryTaskStore struct {
	DB map[string]*task.Task
}

// NewInMemoryTaskStore creates a new InMemoryTaskStore and returns a reference to it.
func NewInMemoryTaskStore() *InMemoryTaskStore {
	return &InMemoryTaskStore{
		DB: make(map[string]*task.Task),
	}
}

// Get retrieves a task from the InMemoryTaskStore and returns it.
func (i *InMemoryTaskStore) Get(key string) (any, error) {
	task, ok := i.DB[key]
	if !ok {
		return nil, fmt.Errorf("failed to find task %s; does it exist?\n", key)
	}
	return task, nil
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

// List creates a slice equal to the number of tasks in the InMemoryTaskStore
// DB, appends each task to the list and returns the list.
func (i *InMemoryTaskStore) List() (any, error) {
	var taskList []*task.Task = make([]*task.Task, len(i.DB))

	for _, t := range i.DB {
		taskList = append(taskList, t)
	}
	return taskList, nil
}

// Count returns the number of tasks in the InMemoryTaskStore DB.
func (i *InMemoryTaskStore) Count() (int, error) {
	return len(i.DB), nil
}
