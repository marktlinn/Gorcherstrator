package scheduler

import (
	"github.com/marktlinn/Gorcherstrator/node"
	"github.com/marktlinn/Gorcherstrator/task"
)

type Epvm struct {
	Name string
}

func (g *Epvm) SelectCandidateNodes(t task.Task, nodes []*node.Node) []*node.Node {
	// TODO:
	return nil
}

func (g *Epvm) Score(t task.Task, nodes []*node.Node) map[string]float64 {
	// TODO:
	return nil
}

func (g *Epvm) Pick(candidates map[string]float64, nodes []*node.Node) *node.Node {
	// TODO:
	return nil
}
