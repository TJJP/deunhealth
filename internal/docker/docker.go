package docker

import (
	"errors"
	"fmt"

	"github.com/moby/moby/client"
)

var _ Dockerer = (*Docker)(nil)

type Dockerer interface {
	UnhealthyGetter
	UnhealthyStreamer
	ContainerRestarter
}

type Docker struct {
	client client.CommonAPIClient
}

var (
	ErrCreateDockerClient = errors.New("cannot create Docker client")
)

func New() (d *Docker, err error) {
	client, err := client.NewClientWithOpts()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCreateDockerClient, err)
	}

	return &Docker{
		client: client,
	}, nil
}
