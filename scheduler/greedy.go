package scheduler

import (
	"github.com/marktlinn/Gorcherstrator/node"
	"github.com/marktlinn/Gorcherstrator/task"
)

type Greedy struct {
	Name string
}

func (g *Greedy) SelectCandidateNodes(t task.Task, nodes []*node.Node) []*node.Node {
	// TODO:
	return nil
}

func (g *Greedy) Score(t task.Task, nodes []*node.Node) map[string]float64 {
	// TODO:
	return nil
}

func (g *Greedy) Pick(candidates map[string]float64, nodes []*node.Node) *node.Node {
	// TODO:
	return nil
}
