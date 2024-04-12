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

// New instantiates a new Manager and returns a pointer to the newly
// instantiated Manager.
func New(workers []string) *Manager {
	taskDB := make(map[uuid.UUID]*task.Task)
	eventDB := make(map[uuid.UUID]*task.TaskEvent)
	taskWorkerMap := make(map[uuid.UUID]string)
	workerTaskMap := make(map[string][]uuid.UUID)

	for w := range workers {
		workerTaskMap[workers[w]] = []uuid.UUID{}
	}

	return &Manager{
		Workers:       workers,
		TaskDB:        taskDB,
		EventDB:       eventDB,
		Pending:       *queue.New(),
		TaskWorkerMap: taskWorkerMap,
		WorkerTaskMap: workerTaskMap,
	}
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

// UpdateTasks intermittently quiries Workers to retrieve their current state.
// Each Worker's current state is updated in the Manager's TaskDB.
func (m *Manager) UpdateTasks() {
	tasks, err := collectTasks(m)
	if err != nil {
		log.Printf("failed to generate slice of tasks: %s\n", err)
	}
	if err := updateCollectedTasks(tasks, m); err != nil {
		log.Printf("failed to update tasks in Manager: %s", err)
	}
}

// AddTask adds Tasks to the Manager's queue.
func (m *Manager) AddTask(te task.TaskEvent) {
	log.Printf("adding task: %+v\n", te)
	m.Pending.Enqueue(te)
}

// collectTasks loops through all the tasks in the Manager's Workers.
// It returns a reference to a slice ot all the tasks found across all Workers.
func collectTasks(m *Manager) ([]*task.Task, error) {
	var tasks []*task.Task

	for _, worker := range m.Workers {
		log.Printf("getting tasks from worker %v\n", worker)
		url := fmt.Sprintf("http://%s/tasks", worker)
		res, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to get tasks from worker %s at url: %s; %s\n",
				worker,
				url,
				err,
			)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to send request: %s\n", err)
		}

		data := json.NewDecoder(res.Body)

		err = data.Decode(&tasks)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshall task data: %s\n", err)
		}
	}
	return tasks, nil
}

// updateCollectedTasks loops through the slice of provided tasks
// and synchronises the the state of the Task with the state of the Task
// of matching ID in the Manager's TaskDB.
func updateCollectedTasks(tasks []*task.Task, m *Manager) error {
	for _, t := range tasks {
		log.Printf("updating tasks...")

		_, ok := m.TaskDB[t.ID]
		if !ok {
			return fmt.Errorf("failed to find task with id: %s\n", t.ID)
		}

		if m.TaskDB[t.ID].State != t.State {
			m.TaskDB[t.ID].State = t.State
		}

		m.TaskDB[t.ID].ContainerID = t.ContainerID
		m.TaskDB[t.ID].StartTime = t.StartTime
		m.TaskDB[t.ID].FinishTime = t.FinishTime
	}
	return nil
}
