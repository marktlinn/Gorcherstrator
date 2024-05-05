package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/marktlinn/Gorcherstrator/manager"
	"github.com/marktlinn/Gorcherstrator/scheduler"
	"github.com/marktlinn/Gorcherstrator/store"
	"github.com/marktlinn/Gorcherstrator/worker"
)

func main() {
	wHost := os.Getenv("WORKER_HOST")
	wPort, _ := strconv.Atoi(os.Getenv("WORKER_PORT"))

	mHost := os.Getenv("MANAGER_HOST")
	mPort, _ := strconv.Atoi(os.Getenv("MANAGER_PORT"))

	fmt.Println("Starting Worker")

	w1 := worker.New("ex_worker1", store.MEMORY)
	workerApi := worker.Api{Address: wHost, Port: wPort, Worker: w1}

	w2 := worker.New("ex_worker2", store.MEMORY)
	workerApi2 := worker.Api{Address: wHost, Port: wPort + 1, Worker: w2}

	w3 := worker.New("ex_worker3", store.MEMORY)
	workerApi3 := worker.Api{Address: wHost, Port: wPort + 2, Worker: w3}

	go w1.RunTasks()
	go w1.CollectStats()
	go w1.UpdateTasks()
	go workerApi.Start()

	go w2.RunTasks()
	go w2.CollectStats()
	go w2.UpdateTasks()
	go workerApi2.Start()

	go w3.RunTasks()
	go w3.CollectStats()
	go w3.UpdateTasks()
	go workerApi3.Start()

	workers := []string{
		fmt.Sprintf("%s:%d", wHost, wPort),
		fmt.Sprintf("%s:%d", wHost, wPort+1),
		fmt.Sprintf("%s:%d", wHost, wPort+2),
	}

	m := manager.New(workers, scheduler.EPVM, store.MEMORY)
	managerApi := manager.Api{Address: mHost, Port: mPort, Manager: m}

	go m.ProcessTasks()
	go m.UpdateTasks()
	go m.RunHealthChecks()

	managerApi.Start()
}
