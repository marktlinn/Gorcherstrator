package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/marktlinn/Gorcherstrator/task"
	"github.com/marktlinn/Gorcherstrator/worker"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

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
