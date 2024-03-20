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
	ContainerID   string
	State         State
	CPU           float64
	Name          string
	Disk          int64
	Memory        int64
	Image         string
	ExposedPorts  nat.PortSet
	RestartPolicy container.RestartPolicyMode
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
	RestartPolicy container.RestartPolicyMode // ["", "always", "unless-stopped", "on-failure"]
	Env           []string
	ExposedPorts  nat.PortSet
}

func NewConfig(t *Task) *Config {
	return &Config{
		Name:          t.Name,
		CPU:           t.CPU,
		Memory:        t.Memory,
		Image:         t.Image,
		Disk:          t.Disk,
		RestartPolicy: t.RestartPolicy,
		ExposedPorts:  t.ExposedPorts,
	}
}

// Docker encapsulates all the data needed to run Tasks inside
// a Docker container.
type Docker struct {
	Client *client.Client
	Config Config
}

// Creates a NewDocker clients with options.
func NewDocker(c *Config) *Docker {
	dc, _ := client.NewClientWithOpts(client.FromEnv)
	return &Docker{
		Client: dc,
		Config: *c,
	}
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

	resp, err := d.Client.ContainerCreate(
		ctx,
		&containerConfig,
		&hostConfig,
		nil,
		nil,
		d.Config.Name,
	)
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

func (d *Docker) Stop(id string) DockerResult {
	log.Printf("Stopping contianer: %v\n", id)

	ctx := context.Background()
	err := d.Client.ContainerStop(ctx, id, container.StopOptions{})
	if err != nil {
		log.Printf("Error: unable to stop container -> %s: %v\n", id, err)
		return DockerResult{Error: err}
	}

	err = d.Client.ContainerRemove(ctx, id, container.RemoveOptions{
		Force:         false,
		RemoveVolumes: true,
		RemoveLinks:   false,
	})
	if err != nil {
		log.Printf("Error: unable to remove container -> %s: %v\n", id, err)
	}

	return DockerResult{Action: "stop", Result: "success", Error: nil}
}
