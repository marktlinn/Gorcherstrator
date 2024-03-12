package main

import (
	"fmt"
	"os"
	"time"

	"github.com/docker/docker/client"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/marktlinn/Gorcherstrator/manager"
	"github.com/marktlinn/Gorcherstrator/node"
	"github.com/marktlinn/Gorcherstrator/task"
	"github.com/marktlinn/Gorcherstrator/worker"
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

func main() {
	t := task.Task{
		ID:     uuid.New(),
		Name:   "example_task",
		State:  task.Pending,
		Disk:   1,
		Memory: 512,
		Image:  "example_image",
	}

	te := task.TaskEvent{
		ID:        uuid.New(),
		Task:      t,
		State:     task.Pending,
		Timestamp: time.Now(),
	}

	w := worker.Worker{
		Name:  "example_worker",
		Queue: *queue.New(),
		DB:    make(map[uuid.UUID]*task.Task),
	}

	m := manager.Manager{
		Pending: *queue.New(),
		Workers: []string{w.Name},
		TaskDB:  make(map[string][]*task.Task),
		EventDB: make(map[string][]*task.TaskEvent),
	}

	n := node.Node{
		Name:   "example_node",
		IP:     "127.0.0.1",
		Memory: 512,
		Disk:   4,
		Role:   "worker",
		Cores:  4,
	}

	fmt.Printf("task: %v\n", t)
	fmt.Printf("task event: %v\n", te)
	fmt.Printf("worker: %v\n", w)

	w.CollectStats()
	w.RunTask()
	w.StartTask()
	w.StopTask()

	fmt.Printf("manager: %v\n", m)
	m.SelectWorker()
	m.UpdateTask()
	m.SendWork()

	fmt.Printf("node: %v\n", n)

	fmt.Println("creating a container...")
	dockerTask, res := createContainer()
	if res.Error != nil {
		fmt.Printf("%v", res.Error)
		os.Exit(1)
	}
	time.Sleep(time.Second * 5)
	stopContainer(dockerTask, res.ContainerID)
}
