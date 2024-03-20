package worker

import (
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
	Queue     queue.Queue
	Name      string
	DB        map[uuid.UUID]*task.Task
	TaskCount int
}

// RunTask method runs a task on the worker.
// RunTask assess the State of the Task.
// If a Task is not started, RunTask starts the Task.
// Else if a Task's state is started, RunTask will stop the task.
func (w *Worker) RunTask() {
	fmt.Println("task running...")
}

// StopTask stops a Task.
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

// CollectStats collects and outputs data about the worker.
func (w *Worker) CollectStats() {
	fmt.Println("stats collected...")
}
