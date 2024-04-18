package worker

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/marktlinn/Gorcherstrator/task"
)

// A Worker has is the layer above a Task.
// It is responsible for:
// - Runing Tasks in Docker containers.
// - Providing stats on runnings Tasks to the manager.
// - Tracking the State of Tasks
// - Accepting instruction from the Manager to run Tasks.
type Worker struct {
	// Queue represents the desired state of Tasks
	Queue queue.Queue
	Name  string
	// DB represents the current actual state of the Tasks.
	DB        map[uuid.UUID]*task.Task
	Stats     *Stats
	TaskCount int
}

// RunTask method runs a task on the worker.
// RunTask assess the State of the Task.
// If a Task is not started, RunTask starts the Task.
// Else if a Task's state is started, RunTask will stop the task.
func (w *Worker) runTask() task.DockerResult {
	t := w.Queue.Dequeue()
	if t == nil {
		log.Println("Queue is empty, no tasks to process")
		return task.DockerResult{Error: nil}
	}

	queuedTask := t.(task.Task)

	taskPersisted := w.DB[queuedTask.ID]
	if taskPersisted == nil {
		taskPersisted = &queuedTask
		w.DB[queuedTask.ID] = &queuedTask
	}

	var result task.DockerResult
	if task.ValidStateTransition(taskPersisted.State, queuedTask.State) {
		switch queuedTask.State {
		case task.Scheduled:
			result = w.StartTask(queuedTask)
		case task.Complete:
			result = w.StopTask(queuedTask)
		default:
			unexpectedError := fmt.Errorf("undefined state of queued task: %+v\n", queuedTask.State).
				Error()
			result.Error = errors.New(unexpectedError)
		}
	} else {
		err := fmt.Errorf("error: transition from %+v to %+v is not valid", taskPersisted.State, queuedTask.State)
		result.Error = err
	}

	return result
}

// StopTask stops a Task which is running on the Worker, gracefully.
func (w *Worker) StopTask(t task.Task) task.DockerResult {
	config := task.NewConfig(&t)
	docker := task.NewDocker(config)

	result := docker.Stop(t.ContainerID)
	if result.Error != nil {
		log.Printf("Error stopping container %v: %v\n",
			t.ContainerID,
			result.Error,
		)
	}
	t.FinishTime = time.Now().UTC()
	t.State = task.Complete
	w.DB[t.ID] = &t
	log.Printf("Container %v stopped and removed for task %v", t.ContainerID, t.ID)
	return result
}

// Starts a task, setting the start time of the task
// and creating a Docker instance to communicate with the Docker
// Daemon.
func (w *Worker) StartTask(t task.Task) task.DockerResult {
	t.StartTime = time.Now().UTC()
	config := task.NewConfig(&t)
	d := task.NewDocker(config)
	result := d.Run()

	if result.Error != nil {
		log.Printf("Error running task %v: %v\n", t.ID, result.Error)
		w.DB[t.ID] = &t
		t.State = task.Failed
		return result
	}

	t.State = task.Running
	t.ContainerID = result.ContainerID
	w.DB[t.ID] = &t

	return result
}

// Enqueues a task to the Worker queue
func (w *Worker) QueueTask(t task.Task) {
	w.Queue.Enqueue(t)
}

// GetTasks returns a lists of Tasks within the Worker's DB.
func (w *Worker) GetTasks() []*task.Task {
	tasks := []*task.Task{}
	for _, t := range w.DB {
		tasks = append(tasks, t)
	}
	return tasks
}

// CollectStats runs GetStats() to maintain an up-to-date collection of stats from a Worker about the Worker and its Tasks.
// Stats are updated once every 15 seconds.
func (w *Worker) CollectStats() {
	for {
		log.Println("Collecting stats")
		w.Stats = GetStats()
		w.Stats.TaskCount = w.TaskCount
		log.Printf("taskCount was: %d\n", w.Stats.TaskCount)
		time.Sleep(15 * time.Second)
	}
}

// RunTasks
func (w *Worker) RunTasks() {
	for {
		if w.Queue.Len() != 0 {
			result := w.runTask()
			if result.Error != nil {
				log.Printf("Error running task: %v\n", result.Error)
			}
		} else {
			log.Printf("No tasks to process currently.\n")
		}
		log.Println("Sleeping for 10 seconds.")
		time.Sleep(10 * time.Second)
	}
}
