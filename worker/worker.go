package worker

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/marktlinn/Gorcherstrator/stats"
	"github.com/marktlinn/Gorcherstrator/store"
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
	DB        store.Store
	Stats     *stats.Stats
	TaskCount int
}

// New creates a new instance of a TaskSTore of the specified DbType
func New(name, dbType string) *Worker {
	w := Worker{
		Name:  name,
		Queue: *queue.New(),
	}

	var s store.Store
	switch dbType {
	case store.MEMORY:
		s = store.NewInMemoryTaskStore()
	default:
		log.Printf("failed to create taskDBStore of type '%s'\n", dbType)
		return nil
	}
	w.DB = s
	return &w
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
		if err := w.DB.Put(t.ID.String(), &t); err != nil {
			log.Printf("failed to insert task %s into store: %s\n", t.ID.String(), err)
		}
		t.State = task.Failed
		return result
	}

	t.State = task.Running
	t.ContainerID = result.ContainerID
	if err := w.DB.Put(t.ID.String(), &t); err != nil {
		log.Printf("failed to insert task %s into store: %s\n", t.ID.String(), err)
	}
	return result
}

// RunTasks pops a queued task from the Worker's queue and runs it.
// RunTasks intermittently (every 10 seconds) checks for tasks on the Worker's queue.
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

// UpdateTasks runs recursively through the Worker to determine if a Task's state is `running` or else `failed`
func (w *Worker) UpdateTasks() {
	var rest time.Duration = 15
	for {
		log.Printf("Updating task status on Worker %s\n", w.Name)
		w.updateTasks()
		log.Printf("Tasks updated for Worker %s; sleeping for %d\n", w.Name, rest)
		time.Sleep(rest * time.Second)
	}
}

// updateTasks is a helper method. The updateTasks method, simply queries thetask's state from the Docker daemon. to determine if the task's state is `running` or not.
// If not the task's state is not `running`, it sets it to `failed`.
func (w *Worker) updateTasks() {
	tasks, err := w.DB.List()
	if err != nil {
		log.Printf("failed to get tasks list: %s\n", err)
		return
	}

	for _, t := range tasks.([]*task.Task) {
		if t.State == task.Running {
			res := w.InspectTask(*t)
			if res.Error != nil {
				log.Printf(
					"failed to inspect task %s in docker container: %s\n", t.ID.String(),
					res.Error,
				)
			}

			if res.Container == nil {
				log.Printf(
					"failed to locate Docker container for task %s in response: %+v\n",
					t.ID.String(),
					res,
				)
				if err := w.DB.Put(t.ID.String(), t); err != nil {
					log.Printf("failed to insert task %s into store: %s\n", t.ID.String(), err)
				}
				t.State = task.Failed
			}

			if res.Container.State.Status == "exited" {
				log.Printf(
					"Container for task %s, not running, found in %s state\n", t.ID.String(),
					res.Container.State.Status,
				)
				if err := w.DB.Put(t.ID.String(), t); err != nil {
					log.Printf("failed to insert task %s into store: %s\n", t.ID.String(), err)
				}
				t.State = task.Failed
			}

			t.HostPorts = res.Container.NetworkSettings.NetworkSettingsBase.Ports
			if err := w.DB.Put(t.ID.String(), t); err != nil {
				log.Printf("failed to insert task %s into store: %s\n", t.ID.String(), err)
			}
		}
	}
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
	if err := w.DB.Put(t.ID.String(), &t); err != nil {
		log.Printf("failed to insert task %s into store: %s\n", t.ID.String(), err)
	}
	log.Printf("Container %v stopped and removed for task %v", t.ContainerID, t.ID)
	return result
}

// InspectTask sets up a task config and Docker instance in order to allow for interaction with the running Docker daemon which is running on the given Worker.
func (w *Worker) InspectTask(t task.Task) task.DockerInspectResponse {
	config := task.NewConfig(&t)
	d := task.NewDocker(config)
	return d.Inspect(t.ContainerID)
}

// Enqueues a task to the Worker queue
func (w *Worker) QueueTask(t task.Task) {
	w.Queue.Enqueue(t)
}

// GetTasks returns a lists of Tasks within the Worker's DB.
func (w *Worker) GetTasks() []*task.Task {
	tasks, err := w.DB.List()
	if err != nil {
		log.Printf("failed to create list of tasks %s\n", err)
		return nil
	}
	return tasks.([]*task.Task)
}

// CollectStats runs GetStats() to maintain an up-to-date collection of stats from a Worker about the Worker and its Tasks.
// Stats are updated once every 15 seconds.
func (w *Worker) CollectStats() {
	for {
		log.Println("Collecting stats")
		w.Stats = stats.GetStats()
		w.Stats.TaskCount = w.TaskCount
		log.Printf("taskCount was: %d\n", w.Stats.TaskCount)
		time.Sleep(15 * time.Second)
	}
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

	taskToPutInDb := t.(task.Task)

	if err := w.DB.Put(taskToPutInDb.ID.String(), &taskToPutInDb); err != nil {
		errMsg := fmt.Errorf("failed to store task %s in DB: %s\n", taskToPutInDb.ID, err)
		log.Println(errMsg)
		return task.DockerResult{Error: errMsg}
	}

	taskGotFromDB, err := w.DB.Get(taskToPutInDb.ID.String())
	if err != nil {
		errMsg := fmt.Errorf("failed to GET task %s from DB: %s\n", taskToPutInDb.ID, err)
		return task.DockerResult{Error: errMsg}
	}

	taskPersisted := *taskGotFromDB.(*task.Task)

	var result task.DockerResult
	if task.ValidStateTransition(taskPersisted.State, taskToPutInDb.State) {
		switch taskToPutInDb.State {
		case task.Scheduled:
			result = w.StartTask(taskToPutInDb)
		case task.Complete:
			result = w.StopTask(taskToPutInDb)
		default:
			unexpectedError := fmt.Errorf("undefined state of queued task: %+v\n", taskToPutInDb.State).
				Error()
			result.Error = errors.New(unexpectedError)
		}
	} else {
		err := fmt.Errorf("error: transition from %+v to %+v is not valid", taskPersisted.State, taskToPutInDb.State)
		result.Error = err
	}

	return result
}
