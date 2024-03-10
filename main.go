package main

import (
	"fmt"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/marktlinn/Gorcherstrator/manager"
	"github.com/marktlinn/Gorcherstrator/node"
	"github.com/marktlinn/Gorcherstrator/task"
	"github.com/marktlinn/Gorcherstrator/worker"
)

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
}
