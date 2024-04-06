package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/marktlinn/Gorcherstrator/task"
	"github.com/marktlinn/Gorcherstrator/worker"

	"github.com/docker/docker/client"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

func createContainer() (*task.Docker, *task.DockerResult) {
	config := task.Config{
		Name:  "exmaple_container",
		Image: "postgres:14",
		Env: []string{
			"POSTGRES_PASSWORD=example_pw",
			"POSTGRES_USER=Gorchestrator",
		},
	}
	dockerClient, _ := client.NewClientWithOpts(client.FromEnv)
	d := task.Docker{
		Client: dockerClient,
		Config: config,
	}

	res := d.Run()
	if res.Error != nil {
		fmt.Printf("%v\n", res.Error)
		return nil, nil
	}

	fmt.Printf("Container %s is up; config set: %v\n", res.ContainerID, config)
	return &d, &res
}

func stopContainer(d *task.Docker, id string) *task.DockerResult {
	res := d.Stop(id)
	if res.Error != nil {
		fmt.Printf("%v\n", res.Error)
		return nil
	}

	fmt.Printf("Container %s stopped and removed\n", res.ContainerID)
	return &res
}

func runTasks(w *worker.Worker) {
	for {
		if w.Queue.Len() != 0 {
			result := w.RunTask()
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

func main() {
	host := os.Getenv("WORKER_HOST")
	port, _ := strconv.Atoi(os.Getenv("WORKER_PORT"))

	fmt.Println("Starting Worker")

	w := worker.Worker{
		Queue: *queue.New(),
		DB:    make(map[uuid.UUID]*task.Task),
	}
	api := worker.Api{Address: host, Port: port, Worker: &w}

	go runTasks(&w)
	go w.CollectStats()
	api.Start()
}
