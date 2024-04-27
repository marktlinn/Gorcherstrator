package scheduler

import (
	"math"

	"github.com/marktlinn/Gorcherstrator/node"
	"github.com/marktlinn/Gorcherstrator/task"
)

// Epvm (Enhanced Parallel Virtual Machine) is a Scheduler that
// assesses the potential "marginal" cost of the new Task with respect to total resources available on the Worker.
// Epvm seeks to spread the computational cost among Workers evenly, in respect to each individual Task's requirements.
type Epvm struct {
	Name string
}

func (g *Epvm) SelectCandidateNodes(t task.Task, nodes []*node.Node) []*node.Node {
	var candidates []*node.Node
	for node := range nodes {
		if checkDisk(t, nodes[node].Disk-nodes[node].DiskAllocated) {
			candidates = append(candidates, nodes[node])
		}
	}

	return candidates
}

func (g *Epvm) Score(t task.Task, nodes []*node.Node) map[string]float64 {
	var lieb_sqr_const float64 = 1.53960071783900203869

	nodesScores := make(map[string]float64)
	maxJobs := 4.0

	for _, node := range nodes {
		cpuUsage, _ := calculateCpuUsage(node)
		cpuLoad := calculateLoad(*cpuUsage, math.Pow(2, 0.8))

		memAllocated := float64(node.Stats.MemUsedKB()) + float64(node.MemoryAllocated)
		memPercentAllocated := memAllocated / float64(node.Memory)
		newMemPercent := (calculateLoad(memAllocated*float64(t.Memory/1000), float64(node.Memory)))

		memCost := math.Pow(
			lieb_sqr_const,
			newMemPercent,
		) + math.Pow(
			lieb_sqr_const,
			(float64(node.TaskCount+1))/maxJobs,
		) - math.Pow(
			lieb_sqr_const,
			memPercentAllocated,
		) - math.Pow(
			lieb_sqr_const,
			float64(node.TaskCount)/float64(maxJobs),
		)

		cpuCost := math.Pow(
			lieb_sqr_const,
			cpuLoad,
		) + math.Pow(
			lieb_sqr_const,
			(float64(node.TaskCount+1))/maxJobs,
		) - math.Pow(
			lieb_sqr_const,
			cpuLoad,
		) - math.Pow(
			lieb_sqr_const,
			float64(node.TaskCount)/float64(maxJobs),
		)

		nodesScores[node.Name] = memCost + cpuCost
	}
	return nodesScores
}

func (g *Epvm) Pick(candidates map[string]float64, nodes []*node.Node) *node.Node {
	// TODO:
	return nil
}
