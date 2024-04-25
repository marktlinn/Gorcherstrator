package scheduler

import (
	"github.com/marktlinn/Gorcherstrator/node"
	"github.com/marktlinn/Gorcherstrator/task"
)

// RoundRobin schedules the tasks evenly across workers, with each task going to a new worker. It starts from the beginning again once all Workers have had tasks assigned.
type RoundRobin struct {
	Name       string
	LastWorker int
}

// SelectCandidateNodes simply returns the nodes that it receives, this implementation is implemented to meet the contract of the Scheduler's interface.
func (r *RoundRobin) SelectCandidateNodes(t task.Task, nodes []*node.Node) []*node.Node {
	return nodes
}

// Score ensures iteration through the nodes in equal order. It iterates over the nodes and ensure the next node on from the previously assign Worker node is scored ready for a Task.
func (r *RoundRobin) Score(t task.Task, nodes []*node.Node) map[string]float64 {
	nodesScores := make(map[string]float64)

	var nextWorker int
	if r.LastWorker+1 < len(nodes) {
		nextWorker = r.LastWorker + 1
		r.LastWorker++
	} else {
		nextWorker = 0
		r.LastWorker = 0
	}

	for i, node := range nodes {
		if i == nextWorker {
			nodesScores[node.Name] = 0.1
		} else {
			nodesScores[node.Name] = 1.0
		}
	}

	return nodesScores
}

// Pick selects the top scoring node from a list of candidate nodes and returns it.
func (r *RoundRobin) Pick(scores map[string]float64, candidates []*node.Node) *node.Node {
	var topScoringNode *node.Node
	var lowestScore float64

	for i, node := range candidates {
		if i == 0 {
			topScoringNode = node
			lowestScore = scores[node.Name]
			continue
		}

		if scores[node.Name] < lowestScore {
			topScoringNode = node
			lowestScore = scores[node.Name]
		}
	}

	return topScoringNode
}
