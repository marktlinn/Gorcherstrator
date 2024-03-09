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
	Pending       queue.Queue
	TaskDB        map[string][]*task.Task
	EventDB       map[string][]*task.TaskEvent
	TaskWorkerMap map[uuid.UUID]string
	WorkerTaskMap map[string][]uuid.UUID
}

func (m *Manager) SelectWorker() {
	fmt.Println("Selecting a Worker...")
}

func (m *Manager) SendWork() {
	fmt.Println("Sending work to Worker...")
}

func (m *Manager) UpdateTask() {
	fmt.Println("Updating Task...")
}
