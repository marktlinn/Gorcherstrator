package task

import (
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
)

type State int

const (
	Pending State = iota
	Running
	Scheduled
	Failed
	Complete
)

// Task represents all the associated properties a task possesses.
type Task struct {
	ID            uuid.UUID
	State         State
	Name          string
	Disk          int
	Memory        int
	ExposedPorts  nat.PortSet
	RestartPolicy string
	portBindings  map[string]string
	StartTime     time.Time
	FinishTime    time.Time
}

type TaskEvent struct {
	ID        uuid.UUID
	State     State
	Task      Task
	Timestamp time.Time
}
