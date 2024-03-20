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

// TODO: Create container and stateValidation Aux funcs
