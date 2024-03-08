package worker

import (
	"fmt"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/marktlinn/Gorcherstrator/task"
)

type worker struct {
	Queue     queue.Queue
	Name      string
	DB        map[uuid.UUID]*task.Task
	TaskCount int
}

// RunTask method runs a task on the worker.
// RunTask assess the State of the Task.
// If a Task is not started, RunTask starts the Task.
// Else if a Task's state is started, RunTask will stop the task.
func (w *worker) RunTask() {
	fmt.Println("task running...")
}

// StartTask starts a Task.
func (w *worker) StartTask() {
	fmt.Println("task started...")
}

// StopTask stops a Task.
func (w *worker) StopTask() {
	fmt.Println("task stopped...")
}

// CollectStats collects and outputs data about the worker.
func (w *worker) CollectStats() {
	fmt.Println("stats collected...")
}
