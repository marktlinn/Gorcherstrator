package scheduler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/marktlinn/Gorcherstrator/node"
	"github.com/marktlinn/Gorcherstrator/task"
	"google.golang.org/grpc/benchmark/stats"
)

// Scheduler determines how to distrubute the workloads/Tasks to Workers.
// The responsibilities of the Scheduler are:
// - Score the workers from best/most appropriate to worst/least appropriate.
// - Pick Worker with the best Score.
// - Select the grouped of Candidate Workers on which a Task can potentially be run.
type Scheduler interface {
	// SelectCandidateNodes filters available workers and returns a slice of appropriate workers that can handle the new Task.
	SelectCandidateNodes(t task.Task, node []*node.Node) []*node.Node
	// Score does the work of assessing the Worker nodes given each a score for the against the Task.
	Score(t task.Task, nodes []*node.Node) map[string]float64
	// Pick is responsible for taking the given candidate node list and their scores and selecting the one with the best score to run the Task.
	Pick(scores map[string]float64, candidates []*node.Node) *node.Node
}

const (
	GREEDY      string = "greedy"
	ROUND_ROBIN string = "roundrobin"
	EVPM        string = "epvm"
)

func SetSchedulerType(schedulerType string) Scheduler {
	switch schedulerType {
	case GREEDY:
		return &Greedy{Name: GREEDY}
	case ROUND_ROBIN:
		return &RoundRobin{Name: ROUND_ROBIN}
	default:
		return &Epvm{Name: EVPM}
	}
}

func checkDisk(t task.Task, diskAvailable int64) bool {
	return t.Disk <= diskAvailable
}

func calculateLoad(usage, capacity float64) float64 {
	return usage / capacity
}

func calculateCpuUsage(node *node.Node) (*float64, error) {
	// TODO:
	return nil, nil
}

func getNodeStats(node *node.Node) *stats.Stats {
	url := fmt.Sprintf("%s/stats", node.Api)
	res, err := http.Get(url)
	if err != nil {
		log.Printf("failed to connect to %s: %s\n", node.Api, err)
	}

	if res.StatusCode != 200 {
		log.Printf(
			"failed to retrieve stats from node: %s; response StatusCode: %d\n",
			node.Api,
			res.StatusCode,
		)
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var status stats.Stats
	json.Unmarshal(body, &status)
	return &status
}
