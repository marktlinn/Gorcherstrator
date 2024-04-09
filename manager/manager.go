package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/marktlinn/Gorcherstrator/task"
	"github.com/marktlinn/Gorcherstrator/worker"
)

// Manager sits above the Worker and tracks of the Workers in the cluster.
// The Manager's responsibilities include:
// - providing the API which allows users to start and stop Tasks.
// - Trackings Tasks, Task State and where they are running in the cluster.
// - Scheduling Tasks on Workers.
type Manager struct {
	Workers       []string
	WorkerTaskMap map[string][]uuid.UUID
	TaskWorkerMap map[uuid.UUID]string
	// LastWorker represents the index of the last Worker in the Workers slice.
	LastWorker int
	Pending    queue.Queue
	// TaskDB holds references to all tasks across all workers.
	TaskDB map[uuid.UUID]*task.Task
	// EventDB holds references to all tasks' metadata.
	EventDB map[uuid.UUID]*task.TaskEvent
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

// SendWork organised the distribution of tasks amongst the Workers and updated the state of the Task.
func (m *Manager) SendWork() {
	if m.Pending.Len() <= 0 {
		log.Println("Queue is empty, no Tasks to process.")
		return
	}
	w := m.SelectWorker()

	t := m.Pending.Dequeue()
	taskEvent := t.(task.TaskEvent)
	tsk := taskEvent.Task

	m.EventDB[taskEvent.ID] = &taskEvent
	m.WorkerTaskMap[w] = append(m.WorkerTaskMap[w], taskEvent.Task.ID)
	m.TaskWorkerMap[tsk.ID] = w

	tsk.State = task.Scheduled
	m.TaskDB[tsk.ID] = &tsk

	data, err := json.Marshal(taskEvent)
	if err != nil {
		log.Printf("failed to marshal task %+v\n", taskEvent)
	}

	url := fmt.Sprintf("http://%s/tasks", w)
	res, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("failed to connect %s; %s\n", url, err)
		m.Pending.Enqueue(taskEvent)
		return
	}

	d := json.NewDecoder(res.Body)
	if res.StatusCode != http.StatusCreated {
		e := worker.ApiErrorResponse{}
		err := d.Decode(&e)
		if err != nil {
			log.Printf("failed to decode response %s\n", err)
			return
		}
		log.Printf("error response: %d; %s\n", e.HTTPStatusCode, e.Message)
		return
	}

	t = task.Task{}
	err = d.Decode(&t)
	if err != nil {
		log.Printf("failed to decode response: %s\n", err)
		return
	}
}

func (m *Manager) UpdateTask() {
	fmt.Println("Updating Task...")
}
