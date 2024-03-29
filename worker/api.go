package worker

import "net/http"

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
