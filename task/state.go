package task

// State represents the state of a task.
// A task can be in one of the following states:
// Pending - task is initialising
// Scheduled - task has been moved to a worker
// Running - task has successfully started running
// Failed - task failed
// Complete - task succeeded, finished and exited without error.
type State int

const (
	Pending State = iota
	Scheduled
	Running
	Failed
	Complete
)

// Creates mappings between the current State (key)
// and the transitional State (values)
var stateTransitions = map[State][]State{
	Pending:   {Scheduled},
	Scheduled: {Scheduled, Failed, Running},
	Running:   {Running, Failed, Complete},
	Failed:    {},
	Complete:  {},
}

// Includes is an auxiliary function to to determine if a given
// state is included within a slice of states
func Includes(states []State, state State) bool {
	for _, s := range states {
		if s == state {
			return true
		}
	}
	return false
}

// checks if a stateTransition is possible form one state to another.
func ValidStateTransition(st State, dest State) bool {
	return Includes(stateTransitions[st], dest)
}
