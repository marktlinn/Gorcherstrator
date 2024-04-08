package manager

import (
	"fmt"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/marktlinn/Gorcherstrator/task"
)

// Manager sits above the Worker and tracks of the Workers in the cluster.
// The Manager's responsibilities include:
// - providing the API which allows users to start and stop Tasks.
// - Trackings Tasks, Task State and where they are running in the cluster.
// - Scheduling Tasks on Workers.
type Manager struct {
	Workers       []string
	WorkerTaskMap map[string][]uuid.UUID
	// LastWorker represents the index of the last Worker in the Workers slice.
	LastWorker int
	Pending    queue.Queue
	// TaskDB holds references to all tasks across all workers.
	TaskDB map[string][]*task.Task
	// EventDB holds references to all tasks' metadata.
	EventDB       map[string][]*task.TaskEvent
	TaskWorkerMap map[uuid.UUID]string
}

// SelectWorker returns a Worker at the index of Manager.LastWorker.
// If LastWorker + 1 is less than the length of Manager.Workers, it is incremented by 1.
// Otherwise it is reset to 0.
func (m *Manager) SelectWorker() string {
	var newWorker int
	if m.LastWorker+1 < len(m.Workers) {
		m.LastWorker++
		newWorker = m.LastWorker
	} else {
		newWorker = 0
		m.LastWorker = 0
	}

	return m.Workers[newWorker]
}

func (m *Manager) SendWork() {
	fmt.Println("Sending work to Worker...")
}

func (m *Manager) UpdateTask() {
	fmt.Println("Updating Task...")
}
