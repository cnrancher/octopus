package exec

import (
	"context"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
)

// DockerContainerStartConfiguration specifies the configuration for starting container.
type DockerContainerStartConfiguration struct {
	Config           *container.Config
	HostConfig       *container.HostConfig
	NetworkingConfig *network.NetworkingConfig
	ContainerName    string
}

// DockerContainerStopConfiguration specifies the configuration for stopping container.
type DockerContainerStopConfiguration struct {
	Cleanup bool
}

type DockerContainer struct {
	containerID string
	autoRemove  bool
}

func (d *DockerContainer) Start(ctx context.Context, config DockerContainerStartConfiguration) error {
	var dClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return errors.Wrap(err, "failed to create docker client")
	}
	defer dClient.Close()

	var createdResp container.ContainerCreateCreatedBody
	for {
		createdResp, err = dClient.ContainerCreate(ctx, config.Config, config.HostConfig, config.NetworkingConfig, nil, config.ContainerName)
		if err != nil {
			if client.IsErrNotFound(err) {
				var resp, err = dClient.ImagePull(ctx, config.Config.Image, types.ImagePullOptions{})
				if err != nil {
					return errors.Wrapf(err, "failed to pull image %s", config.Config.Image)
				}
				_, _ = io.Copy(ioutil.Discard, resp)
				_ = resp.Close()
				continue
			}
			return errors.Wrap(err, "failed to create container")
		}
		break
	}
	d.containerID = createdResp.ID
	d.autoRemove = config.HostConfig.AutoRemove

	err = dClient.ContainerStart(ctx, d.containerID, types.ContainerStartOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to start container")
	}

	if config.Config.Healthcheck != nil {
		err = wait.Poll(1*time.Second, 30*time.Second, func() (bool, error) {
			var inspectedResp, err = dClient.ContainerInspect(ctx, d.containerID)
			if err != nil {
				return false, errors.Wrapf(err, "failed to inspect container")
			}
			if state := inspectedResp.State; state != nil {
				if health := state.Health; health != nil {
					return strings.ToLower(health.Status) == "healthy", nil
				}
			}
			return false, nil
		})
		if err != nil {
			return errors.Wrapf(err, "failed to inspect container")
		}
	}

	return nil
}

func (d *DockerContainer) Stop(ctx context.Context, config DockerContainerStopConfiguration) error {
	if d.containerID == "" {
		return nil
	}

	var dClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return errors.Wrap(err, "failed to create docker client")
	}
	defer dClient.Close()

	err = dClient.ContainerStop(ctx, d.containerID, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to stop container %s", d.containerID)
	}

	if !d.autoRemove {
		err = dClient.ContainerRemove(ctx, d.containerID, types.ContainerRemoveOptions{RemoveVolumes: config.Cleanup, Force: config.Cleanup})
		if err != nil {
			return errors.Wrapf(err, "failed to remove container %s", d.containerID)
		}
	}

	return nil
}

func NewContainer() *DockerContainer {
	return &DockerContainer{}
}
