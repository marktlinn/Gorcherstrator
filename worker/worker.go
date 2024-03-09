package worker

import (
	"fmt"

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
func (w *Worker) StopTask() {
	fmt.Println("task stopped...")
}

// CollectStats collects and outputs data about the worker.
func (w *Worker) CollectStats() {
	fmt.Println("stats collected...")
}
