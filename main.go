package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/marktlinn/Gorcherstrator/manager"
	"github.com/marktlinn/Gorcherstrator/task"
	"github.com/marktlinn/Gorcherstrator/worker"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

func main() {
	wHost := os.Getenv("WORKER_HOST")
	wPort, _ := strconv.Atoi(os.Getenv("WORKER_PORT"))

	mHost := os.Getenv("MANAGER_HOST")
	mPort, _ := strconv.Atoi(os.Getenv("MANAGER_PORT"))

	fmt.Println("Starting Worker")

	w := worker.Worker{
		Queue: *queue.New(),
		DB:    make(map[uuid.UUID]*task.Task),
	}
	workerApi := worker.Api{Address: wHost, Port: wPort, Worker: &w}

	go w.RunTasks()
	go w.CollectStats()
	go w.UpdateTasks()
	go workerApi.Start()

	workers := []string{fmt.Sprintf("%s:%d", wHost, wPort)}

	m := manager.New(workers)
	managerApi := manager.Api{Address: mHost, Port: mPort, Manager: m}

	go m.ProcessTasks()
	go m.UpdateTasks()
	go m.RunHealthChecks()

	managerApi.Start()
}
