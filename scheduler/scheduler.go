package scheduler

import (
	"time"

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

// checkDisk is a simple helper function to determine if the Task's disk requirements are less than is available for a Node.
func checkDisk(t task.Task, diskAvailable int64) bool {
	return t.Disk <= diskAvailable
}

// calculateLoad is a simple helper function determining the CPU load.
func calculateLoad(usage, capacity float64) float64 {
	return usage / capacity
}

// calculateCpuUsage is a helper function that determines the CPU usage of the Linux system running Tasks.
func calculateCpuUsage(node *node.Node) (*float64, error) {
	stat1, err := node.GetStats()
	if err != nil {
		return nil, err
	}

	time.Sleep(3 * time.Second)
	stat2, err := node.GetStats()
	if err != nil {
		return nil, err
	}

	stat1Idle := stat1.CPUStats.User + stat1.CPUStats.IOWait
	stat2Idle := stat2.CPUStats.User + stat2.CPUStats.IOWait
	stat1NonIdle := stat1.CPUStats.User + stat1.CPUStats.Nice + stat1.CPUStats.System + stat1.CPUStats.IRQ + stat1.CPUStats.SoftIRQ + stat1.CPUStats.Steal
	stat2NonIdle := stat2.CPUStats.User + stat2.CPUStats.Nice + stat2.CPUStats.System + stat2.CPUStats.IRQ + stat2.CPUStats.SoftIRQ + stat2.CPUStats.Steal

	stat1Total := stat1Idle + stat1NonIdle
	stat2Total := stat2Idle + stat2NonIdle

	ttl := stat2Total - stat1Total
	ttlIdle := stat2Idle - stat1Idle

	var cpuPercentUsage float64
	if ttl == 0 && ttlIdle == 0 {
		cpuPercentUsage = 0.00
	} else {
		cpuPercentUsage = (float64(ttl) - float64(ttlIdle)) / float64(ttl)
	}
	return &cpuPercentUsage, nil
}
