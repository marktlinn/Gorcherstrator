package store

import (
	"fmt"

	"github.com/marktlinn/Gorcherstrator/task"
)

// InMemoryEventStore is a map structure store of TaskEvents held in memory.
type InMemoryEventStore struct {
	DB map[string]*task.TaskEvent
}

// InMemoryEventStore creates a new NewInMemoryEventStore and returns a reference to it.
func NewInMemoryEventStore() *InMemoryEventStore {
	return &InMemoryEventStore{
		DB: make(map[string]*task.TaskEvent),
	}
}

// Get retrieves a taskEvent from the InMemoryEventStore and returns it.
func (i *InMemoryEventStore) Get(key string) (any, error) {
	event, ok := i.DB[key]
	if !ok {
		return nil, fmt.Errorf("failed to find task %s; does it exists?\n", key)
	}
	return event, nil
}

// Put inserts a key value pair into the NewInMemoryEventStore, asserting first that the value is a pointer to a task.TaskEvent.
func (i *InMemoryEventStore) Put(key string, value any) error {
	event, ok := value.(*task.TaskEvent)
	if !ok {
		return fmt.Errorf("failed to assert value %s as type *task.Task\n", value)
	}
	i.DB[key] = event
	return nil
}

// List creates a slice equal to the number of task.TaskEvents in the InMemoryEventStore
// DB, appends each taskEvent to the list and returns the list.
func (i *InMemoryEventStore) List() (any, error) {
	var eventList []*task.TaskEvent = make([]*task.TaskEvent, len(i.DB))

	for _, t := range i.DB {
		eventList = append(eventList, t)
	}
	return eventList, nil
}

// Count returns the number of taskEvents in the InMemoryEventStore DB.
func (i *InMemoryEventStore) Count() (int, error) {
	return len(i.DB), nil
}
