package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/marktlinn/Gorcherstrator/task"
)

// StartTaskHandler handles requests to initiate a new task. Extracts task details from a JSON-encoded
// 'task.TaskEvent' in the request body, adds the task to the worker queue, and returns the task information to the client.
func (a *Api) StartTaskHandler(w http.ResponseWriter, r *http.Request) {
	data := json.NewDecoder(r.Body)
	data.DisallowUnknownFields()

	taskEvent := task.TaskEvent{}
	if err := data.Decode(&taskEvent); err != nil {
		msg := fmt.Sprintf("Error unmarshalling request body : %v\n", err)
		log.Println(msg)
		e := ApiErrorResponse{
			HTTPStatusCode: 400,
			Message:        msg,
		}
		if err := json.NewEncoder(w).Encode(e); err != nil {
			log.Printf("failed to encode response to json: %s\n", err)
		}
		return
	}

	a.Worker.QueueTask(taskEvent.Task)
	log.Printf("Task %s added to worker %s task queue", taskEvent.ID, a.Worker.Name)
	w.WriteHeader(201)
	if err := json.NewEncoder(w).Encode(taskEvent.Task); err != nil {
		log.Printf("failed to encode response to json: %s\n", err)
	}
}

// GetTaskHandler handles requests to retrieve tasks from the Worker. Returns a JSON-encoded list of tasks currently managed by the worker.
func (a *Api) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	if err := json.NewEncoder(w).Encode(a.Worker.GetTasks()); err != nil {
		log.Printf("failed to encode response to json: %s\n", err)
	}
}

// Handles requests to stop a running task. Takes a taskID from the request path,
// verifies its existence, and adds a copy of the task with a 'Complete' state to
// the worker's queue. This signals the worker to gracefully stop the original task.
func (a *Api) StopTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("taskID")
	fmt.Printf("taskID is %s\n", taskID)
	if taskID == "" {
		log.Printf("TaskID not found in request\n")
		w.WriteHeader(400)
	}

	taskUUID, _ := uuid.Parse(taskID)
	fmt.Printf("UUID found: %v\n", taskUUID)
	targetTask, err := a.Worker.DB.Get(taskUUID.String())
	if err != nil {
		log.Printf("No task matches task ID %v\n", taskUUID)
		w.WriteHeader(404)
	}

	fmt.Printf("TargetTask: %+v\n", targetTask)
	copiedTask := *targetTask.(*task.Task)
	fmt.Printf("copiedTask: %+v\n", copiedTask)
	copiedTask.State = task.Complete
	a.Worker.QueueTask(copiedTask)

	log.Printf("Task %v added to worker %s stop Queue", copiedTask.ID.String(), a.Worker.Name)
	w.WriteHeader(204)
}

// GetStatsHandler provides the api for retrieving the current Stats from a Worker.
// Worker stats are updated every 15 seconds.
func (a *Api) GetStatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	if err := json.NewEncoder(w).Encode(a.Worker.Stats); err != nil {
		log.Printf("failed to get stats: %s\n", err)
	}
}

func (a *Api) InspectTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("taskID")
	if taskID == "" {
		log.Printf("No taskID passed in request.\n")
		w.WriteHeader(400)
	}

	taskUUID, _ := uuid.Parse(taskID)
	t, err := a.Worker.DB.Get(taskUUID.String())
	if err != nil {
		log.Printf("No task with ID %s found: %s\n", taskUUID.String(), err)
		w.WriteHeader(404)
		return
	}

	res := a.Worker.InspectTask(t.(task.Task))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	if err := json.NewEncoder(w).Encode(res.Container); err != nil {
		log.Printf("failed to encode response: %s\n", err)
	}
}
