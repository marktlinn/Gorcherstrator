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

// StartTask starts a Task.
func (w *Worker) StartTask() {
	fmt.Println("task started...")
}

// StopTask stops a Task.
func (w *Worker) StopTask(t task.Task) task.DockerResult {
	config := task.NewConfig(&t)
	docker := task.NewDocker(config)

	result := docker.Stop(t.ContainerId)
	if result.Error != nil {
		log.Printf("Error stopping container %v: %v\n",
			t.ContainerId,
			result.Error,
		)
	}
	t.FinishTime = time.Now().UTC()
	t.State = task.Complete
	w.DB[t.ID] = &t
	log.Printf("Container %v stopped and removed for task %v", t.ContainerId, t.ID)
	return result
}

// CollectStats collects and outputs data about the worker.
func (w *Worker) CollectStats() {
	fmt.Println("stats collected...")
}
