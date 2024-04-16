package manager

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/marktlinn/Gorcherstrator/task"
)

// StartTaskHandler handles requests to initiate a new task. It extracts task details from a JSON-encoded
// 'task.TaskEvent' in the request body, adds the task to the Manager's queue, and returns the task information to the client.
func (a *Api) StartTaskHandler(w http.ResponseWriter, r *http.Request) {
	data := json.NewDecoder(r.Body)
	data.DisallowUnknownFields()
	defer r.Body.Close()

	taskEvent := task.TaskEvent{}
	if err := data.Decode(&taskEvent); err != nil {
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
	// TODO: FINISHED
}

// GetTasks is a helper function which constructs and returns a slice of
// pointers to the tasks in the Manager's DB.
func (m *Manager) GetTasks() []*task.Task {
	tasks := []*task.Task{}
	for _, t := range m.TaskDB {
		tasks = append(tasks, t)
	}
	return tasks
}