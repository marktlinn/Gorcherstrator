package task

import (
	"context"
	"io"
	"log"
	"math"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
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

// Task is the most foundational element fo the Gorchestrator.
// A Task sits at the lowest point, under the Worker, it
// represents all the associated properties a task possesses.
type Task struct {
	ID            uuid.UUID
	State         State
	Name          string
	Disk          int
	Memory        int
	Image         string
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

// The Config for Docker containers
type Config struct {
	Name          string
	AttachStderr  bool
	AttachStdin   bool
	AttachStdout  bool
	CMD           []string
	Memory        int64
	CPU           float64
	Image         string
	Disk          int64
	RestartPolicy string // ["", "always", "unless-stopped", "on-failure"]
	Env           []string
	ExposedPorts  nat.PortSet
}

// Docker encapsulates all the data needed to run Tasks inside
// a Docker container.
type Docker struct {
	Client *client.Client
	Config Config
}

// DockerResult provides an API wrapper for the interactions with a Docker Container.
type DockerResult struct {
	ContainerID string
	Action      string
	Result      string
	Error       error
}

func (d *Docker) Run() DockerResult {
	ctx := context.Background()
	reader, err := d.Client.ImagePull(
		ctx, d.Config.Image, types.ImagePullOptions{},
	)
	if err != nil {
		log.Printf("Error, unable to pull image-> %s: %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}
	io.Copy(os.Stdout, reader)

	restartPolicy := container.RestartPolicy{
		// todo: fix typing
		Name: d.Config.RestartPolicy,
	}

	resources := container.Resources{
		Memory:   d.Config.Memory,
		NanoCPUs: int64(d.Config.CPU * math.Pow(10, 9)),
	}

	hostConfig := container.HostConfig{
		RestartPolicy:   restartPolicy,
		Resources:       resources,
		PublishAllPorts: true,
	}

	containerConfig := container.Config{
		Image:        d.Config.Image,
		Tty:          false,
		Env:          d.Config.Env,
		ExposedPorts: d.Config.ExposedPorts,
	}

	resp, err := d.Client.ContainerCreate(ctx, &containerConfig, &hostConfig, nil, nil, d.Config.Name)
	if err != nil {
		log.Printf("Erro starting Docker Container -> %s: %v", d.Config.Image, err)
		return DockerResult{Error: err}
	}

	out, err := d.Client.ContainerLogs(
		ctx,
		resp.ID,
		container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
		})

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	return DockerResult{ContainerID: resp.ID, Action: "start", Result: "success"}
}
