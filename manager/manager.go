package manager

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/marktlinn/Gorcherstrator/node"
	"github.com/marktlinn/Gorcherstrator/scheduler"
	"github.com/marktlinn/Gorcherstrator/store"
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
	// TaskDB holds references to all Tasks data in a datastore.
	TaskDB store.Store
	// EventDB holds references to Tasks' metadata in a datastore.
	EventDB store.Store
	// A slice of worker nodes.
	WorkerNodes []*node.Node
	// The Scheduler type to be used for scheduling Tasks.
	Scheduler scheduler.Scheduler
}

// New instantiates a new Manager and returns a pointer to the newly
// instantiated Manager.
func New(workers []string, schedulerType, dbType string) *Manager {
	taskWorkerMap := make(map[uuid.UUID]string)
	workerTaskMap := make(map[string][]uuid.UUID)

	var nodes []*node.Node
	for w := range workers {
		workerTaskMap[workers[w]] = []uuid.UUID{}

		nodeApi := fmt.Sprintf("http://%v", workers[w])
		newNode := node.NewNode(workers[w], nodeApi, "worker")
		nodes = append(nodes, newNode)
	}

	s := scheduler.SetSchedulerType(schedulerType)

	m := Manager{
		Workers:       workers,
		Pending:       *queue.New(),
		TaskWorkerMap: taskWorkerMap,
		WorkerTaskMap: workerTaskMap,
		Scheduler:     s,
		WorkerNodes:   nodes,
	}

	var taskStore store.Store
	var eventStore store.Store
	switch dbType {
	case store.MEMORY:
		taskStore = store.NewInMemoryTaskStore()
		eventStore = store.NewInMemoryEventStore()
	}

	m.TaskDB = taskStore
	m.EventDB = eventStore
	return &m
}

// SelectWorker makes use of the Scheduler interface to to nominate an appropriate Worker to receive a Task. If no Worker is found, or no appropriate candidates are given an error is returned.
func (m *Manager) SelectWorker(t task.Task) (*node.Node, error) {
	candidates := m.Scheduler.SelectCandidateNodes(t, m.WorkerNodes)
	if candidates == nil {
		errMsg := fmt.Sprintf("failed to find available candidates for task %s\n", t.ID)
		return nil, errors.New(errMsg)
	}

	nodeScores := m.Scheduler.Score(t, candidates)
	selectedNode := m.Scheduler.Pick(nodeScores, candidates)

	return selectedNode, nil
}

// SendWork organises the distribution of Tasks amongst the Workers and updates the state of the Task.
func (m *Manager) SendWork() {
	if m.Pending.Len() <= 0 {
		log.Println("Queue is empty, no Tasks to process.")
		return
	}

	t := m.Pending.Dequeue()
	taskEvent := t.(task.TaskEvent)

	err := m.EventDB.Put(taskEvent.ID.String(), &taskEvent)
	if err != nil {
		log.Printf("failed to insert task into store %s: %s\n", taskEvent.ID, err)
	}

	taskWorker, ok := m.TaskWorkerMap[taskEvent.Task.ID]
	if ok {
		res, err := m.TaskDB.Get(taskEvent.Task.ID.String())
		if err != nil {
			log.Printf("failed to schedule task: %s\n", err)
			return
		}
		persistedTask, ok := res.(*task.Task)
		if !ok {
			log.Printf("failed to convert task %v to type task.Task\n", res)
			return
		}

		if taskEvent.State == task.Complete &&
			task.ValidStateTransition(persistedTask.State, taskEvent.State) {
			m.stopTask(taskWorker, taskEvent.Task.ID.String())
			return
		}
	}

	tsk := taskEvent.Task
	w, err := m.SelectWorker(tsk)
	if err != nil {
		log.Printf("failed to select Worker for task %s: %s\n", taskEvent.ID, err)
		return
	}
	m.WorkerTaskMap[w.Name] = append(m.WorkerTaskMap[w.Name], taskEvent.Task.ID)
	m.TaskWorkerMap[tsk.ID] = w.Name

	tsk.State = task.Scheduled
	if putErr := m.TaskDB.Put(tsk.ID.String(), &tsk); putErr != nil {
		log.Printf("failed to put task %s in taskDB: %s\n", tsk.ID, putErr)
	}

	data, marshalErr := json.Marshal(taskEvent)
	if marshalErr != nil {
		log.Printf("failed to marshal task %+v\n", taskEvent)
	}

	url := fmt.Sprintf("http://%s/tasks", w.Name)
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
	var rest time.Duration = 15
	for {
		log.Println("checking for task updates in Workers")
		m.updateTasks()
		log.Printf("Tasks updated; resuming in %d seconds\n", rest)
		time.Sleep(rest * time.Second)
	}
}

// updateTasks is a helper function that gets all tasks from each Worker, then ensures the state of each Task
// is in sync with the TaskDB store.
func (m *Manager) updateTasks() {
	for _, worker := range m.Workers {
		log.Printf("getting tasks from worker %v\n", worker)
		url := fmt.Sprintf("http://%s/tasks", worker)
		res, err := http.Get(url)
		if err != nil {
			fmt.Printf(
				"failed to get tasks from worker %s at url: %s; %s\n",
				worker,
				url,
				err,
			)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			fmt.Printf("failed to send request: %s\n", err)
		}

		data := json.NewDecoder(res.Body)
		var t []*task.Task
		err = data.Decode(&t)
		if err != nil && errors.Is(err, io.EOF) {
			fmt.Printf("failed to unmarshall task data: %s\n", err)
		}
		updateCollectedTasks(t, m)
	}
}

// stopTask is a helper function helping connect to the correct Worker where a Task is running and scheduling for that Task to be gracefully terminated.
func (m *Manager) stopTask(worker, taskID string) {
	client := &http.Client{}
	url := fmt.Sprintf("http://%s/tasks/%s", worker, taskID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Printf("failed to create deletion request for task %s\n", taskID)
		return
	}

	res, err := client.Do(req)
	if err != nil {
		log.Printf("failed to connect to Worker %s at %s\n", worker, url)
	}

	if res.StatusCode != 204 {
		log.Printf("failed to send request: %s\n", err)
		return
	}

	log.Printf("task %s successfully scheduled to be stopped\n", taskID)
}

// ProcessTasks, at the determined interval, processes the work on the Manager's queue.
func (m *Manager) ProcessTasks() {
	var rest time.Duration = 10
	for {
		log.Println("Processing tasks in Manager queue")
		m.SendWork()
		log.Printf("Processing complete; resuming in %d seconds\n", rest)
		time.Sleep(rest * time.Second)
	}
}

// AddTask adds Tasks to the Manager's queue.
func (m *Manager) AddTask(te task.TaskEvent) {
	log.Printf("adding task: %+v\n", te)
	m.Pending.Enqueue(te)
}

// updateCollectedTasks loops through the slice of provided tasks
// and synchronises the the state of the Task with the state of the Task
// of matching ID in the Manager's TaskDB.
func updateCollectedTasks(tasks []*task.Task, m *Manager) {
	for _, t := range tasks {
		log.Printf("updating tasks...")
		res, err := m.TaskDB.Get(t.ID.String())
		if err != nil {
			log.Printf("failed to get task for Manager %s\n", err)
			continue
		}

		taskPersisted, ok := res.(*task.Task)
		if !ok {
			log.Printf("failed to convert %v to type task.Task\n", res)
		}
		if taskPersisted.State != t.State {
			taskPersisted.State = t.State
		}

		taskPersisted.StartTime = t.StartTime
		taskPersisted.FinishTime = t.FinishTime
		taskPersisted.ContainerID = t.ContainerID
		taskPersisted.HostPorts = t.HostPorts

		if putErr := m.TaskDB.Put(taskPersisted.ID.String(), taskPersisted); putErr != nil {
			log.Printf("failed to put task %s in taskDB: %s\n", taskPersisted.ID, putErr)
		}
	}
}

// RunHealthChecks ensures running tasks are pinged at a setinterval to ensure they are running correctly.
// A task found to be in any State other than `Running` will be restarted.
// Tasks that repeatedly fail (3 or more times) will be set to a `failed` state.
func (m *Manager) RunHealthChecks() {
	var rest time.Duration = 60
	for {
		m.runHealthCheck()
		log.Printf("HealthChecks complete, next cycle will start in %d seconds.\n", rest)
		time.Sleep(rest * time.Second)
	}
}

// runHealthCheck loops over all Tasks on the Worker performing HealthChecks.
// If a task is not in a `Running` state and its RestartCount is less that 3,
// and attempt will be made to restarted it. Else the Task's State will be set to False to indicate it has failed.
func (m *Manager) runHealthCheck() {
	for _, t := range m.GetTasks() {
		if t.State == task.Running && t.RestartCount < 3 {
			if err := m.healthCheckTask(*t); err != nil {
				m.restartTask(t)
				return
			}
			if t.State == task.Failed {
				m.restartTask(t)
			}
		}
	}
}

// healthCheckTask pulls the IP address and port of the target Task from TaskWorkerMap. It then pings the selected Task's HealthCheck endpoint.
func (m *Manager) healthCheckTask(t task.Task) error {
	log.Printf("Performing HealtCheck on Task %+v\n", t)

	wTask := m.TaskWorkerMap[t.ID]
	hostPort := getHostPort(t.HostPorts)
	if hostPort == nil {
		return fmt.Errorf("hostPort is nil")
	}
	wrkr := strings.Split(wTask, ":")
	url := fmt.Sprintf("http://%s:%s%s", wrkr[0], *hostPort, t.HealthCheck)

	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to connect for healthcheck on task %s: %w", t.ID, err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"failed HealthCheck on task %s, with StatusCode %d\n",
			t.ID,
			res.StatusCode,
		)
	}

	log.Printf("HealthCheck for task %s complete successfully\n", t.ID)
	return nil
}

// getPorts is a auxiliary function to retrieve the a HostPort from a selected Task.
func getHostPort(ports nat.PortMap) *string {
	for k := range ports {
		return &ports[k][0].HostPort
	}
	return nil
}

// restartTasks targets the give task and attempts to restart it, outputting logs for any errors that occur while attempting the restart.
func (m *Manager) restartTask(t *task.Task) {
	wTask := m.TaskWorkerMap[t.ID]
	t.State = task.Scheduled
	t.RestartCount++

	if putErr := m.TaskDB.Put(t.ID.String(), t); putErr != nil {
		log.Printf("failed to put task %s in taskDB: %s\n", t.ID, putErr)
	}
	taskEvent := task.TaskEvent{
		ID:        uuid.New(),
		Task:      *t,
		State:     task.Running,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(taskEvent)
	if err != nil {
		log.Printf("failed to marshal taskEvent data: %+v\n", taskEvent)
	}

	url := fmt.Sprintf("http://%s/tasks", wTask)
	res, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		m.Pending.Enqueue(t)
		log.Printf(
			"failed to connect to Worker for task: %s, requeuing task; err: %s\n",
			wTask,
			err,
		)
		return
	}

	d := json.NewDecoder(res.Body)
	if res.StatusCode != http.StatusCreated {
		e := worker.ApiErrorResponse{}
		err := d.Decode(&e)
		if err != nil {
			fmt.Printf("failed to decode response %s\n", err.Error())
			return
		}
		log.Printf(
			"failed to create task, unexpected HTTP Status %d received: %s\n",
			res.StatusCode,
			err,
		)

		return
	}

	newTask := task.Task{}
	err = d.Decode(&newTask)
	if err != nil {
		fmt.Printf("failed to decode New Task %+v: %s\n", newTask, err)
	}
}
