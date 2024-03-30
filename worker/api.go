package worker

import (
	"fmt"
	"net/http"
)

type ApiErrorResponse struct {
	HTTPStatusCode int
	Message        string
}

// The Api wraps the Worker and exposes its core functionality to the Manager.
type Api struct {
	Address string
	Router  *http.ServeMux
	Worker  *Worker
	Port    int
}

// initRouter initialises the Api Router setting up the necessary routes in the process.
func (a *Api) initRouter() {
	router := http.NewServeMux()
	router.HandleFunc("POST /tasks", a.StartTaskHandler)
	router.HandleFunc("GET /tasks", a.GetTaskHandler)
	router.HandleFunc("DELETE /tasks/{taskID}", a.StartTaskHandler)
}

// Starts the server and invokes the initRouter ensuring the routes are established.
func (a *Api) Start() {
	a.initRouter()
	http.ListenAndServe(fmt.Sprintf("%s:%d", a.Address, a.Port), a.Router)
}
