package scheduler

// Scheduler determines how to distrubute the workloads/Tasks to Workers.
// The responsibilities of the Scheduler are:
// - Score the workers from best/most appropriate to worst/least appropriate.
// - Pick Worker with the best Score.
// - Select the groupd of Candidate Workers on which a Task can potentially be run.
type Scheduler interface {
	Score()
	Pick()
	SelectCandidateNodes()
}
