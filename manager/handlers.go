package manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/marktlinn/Gorcherstrator/task"
)

// StartTaskHandler handles requests to initiate a new task. It extracts task details from a JSON-encoded
// 'task.TaskEvent' in the request body, adds the task to the Manager's queue, and returns the task information to the client.
func (a *Api) StartTaskHandler(w http.ResponseWriter, r *http.Request) {
	data := json.NewDecoder(r.Body)
	data.DisallowUnknownFields()
	defer r.Body.Close()

	taskEvent := task.TaskEvent{}
	if err := data.Decode(&taskEvent); err != nil && !errors.Is(err, io.EOF) {
		errMsg := fmt.Sprintf("failed to unmarshall json body data %s\n", err)
		log.Println(errMsg)
		w.WriteHeader(400)
		errRes := ApiErrorResponse{
			Message:        errMsg,
			HTTPStatusCode: 400,
		}
		if err := json.NewEncoder(w).Encode(errRes); err != nil {
			log.Printf("error encoding json response: %s\n", err)
		}
		return
	}

	a.Manager.AddTask(taskEvent)
	w.WriteHeader(201)
	if err := json.NewEncoder(w).Encode(taskEvent.Task); err != nil {
		log.Printf("error encoding json response: %s\n", err)
	}
}

// GetTaskHandler handles requests to retrieve tasks from the Manager.
// It returns a JSON-encoded list of tasks currently managed by the Manager.
func (a *Api) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	if err := json.NewEncoder(w).Encode(a.Manager.GetTasks()); err != nil {
		log.Printf("error encoding json response: %s\n", err)
	}
}

// StopTaskHandler handles requests to stop a running task. Takes a taskID from the request path,
// verifies its existence, and adds a copy of the task with a 'Complete' state to
// the Manager's queue. This signals the Manager to gracefully stop the task.
func (a *Api) StopTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("taskID")
	if taskID == "" {
		log.Println("No taskID received")
		w.WriteHeader(400)
	}

	tID, _ := uuid.Parse(taskID)
	targetTask, err := a.Manager.TaskDB.Get(tID.String())
	if err != nil {
		log.Printf("Failed to find Task with ID: %s\n", tID)
		w.WriteHeader(404)
		return
	}

	taskEvent := task.TaskEvent{
		ID:        uuid.New(),
		State:     task.Complete,
		Timestamp: time.Now(),
	}

	cpyTargetTask := targetTask.(*task.Task)
	taskEvent.Task = *cpyTargetTask
	a.Manager.AddTask(taskEvent)

	log.Printf("Task %v added to Manager's stop Queue\n", cpyTargetTask.ID)
	w.WriteHeader(204)
}

// GetTasks is a helper function which constructs and returns a slice of
// pointers to the tasks in the Manager's DB.
func (m *Manager) GetTasks() []*task.Task {
	tasksList, err := m.TaskDB.List()
	if err != nil {
		log.Printf("failed to get list of tasks in %v\n", err)
	}

	return tasksList.([]*task.Task)
}
