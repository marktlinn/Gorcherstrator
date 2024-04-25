package scheduler

import (
	"github.com/marktlinn/Gorcherstrator/node"
	"github.com/marktlinn/Gorcherstrator/task"
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
