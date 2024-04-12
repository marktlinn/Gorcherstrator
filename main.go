package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/marktlinn/Gorcherstrator/manager"
	"github.com/marktlinn/Gorcherstrator/task"
	"github.com/marktlinn/Gorcherstrator/worker"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

func main() {
	host := os.Getenv("GORCH_HOST")
	port, _ := strconv.Atoi(os.Getenv("GORCH_PORT"))

	fmt.Println("Starting Worker")

	w := worker.Worker{
		Queue: *queue.New(),
		DB:    make(map[uuid.UUID]*task.Task),
	}
	api := worker.Api{Address: host, Port: port, Worker: &w}

	go runTasks(&w)
	go w.CollectStats()
	go api.Start()

	workers := []string{fmt.Sprintf("%s:%d", host, port)}
	m := manager.New(workers)

	// temporary tasks to demo assigning tasks to Manager.
	for i := 0; i < 3; i++ {
		t := task.Task{
			ID:    uuid.New(),
			State: task.Scheduled,
			Name:  fmt.Sprintf("test-container-%d", i),
			Image: "strm/helloworld-http",
		}
		te := task.TaskEvent{
			ID:    uuid.New(),
			Task:  t,
			State: task.Running,
		}
		log.Printf("task no %d created\n", i)
		m.AddTask(te)
		m.SendWork()
	}

	go func() {
		for {
			fmt.Printf("Manager updating tasks on %d Workers", len(m.Workers))
			m.UpdateTasks()
			time.Sleep(15 * time.Second)
		}
	}()
	for {
		for _, t := range m.TaskDB {
			fmt.Printf("Manager's task %d, State: %d", t.ID, t.State)
			time.Sleep(15 * time.Second)
		}
	}
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
