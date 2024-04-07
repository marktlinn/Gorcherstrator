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
	a.Router = http.NewServeMux()
	a.Router.HandleFunc("POST /tasks", a.StartTaskHandler)
	a.Router.HandleFunc("GET /tasks", a.GetTaskHandler)
	a.Router.HandleFunc("DELETE /tasks/{taskID}", a.StopTaskHandler)
	a.Router.HandleFunc("GET /stats", a.GetStatsHandler)
}

// Starts the server and invokes the initRouter ensuring the routes are established.
func (a *Api) Start() {
	a.initRouter()
	http.ListenAndServe(fmt.Sprintf("%s:%d", a.Address, a.Port), a.Router)
}
