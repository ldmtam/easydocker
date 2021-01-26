package easydocker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
)

var (
	// ErrContainerNotFound container not found
	ErrContainerNotFound = errors.New("container not found")

	// ErrRepositoryEmpty repository is empty
	ErrRepositoryEmpty = errors.New("repository can't be empty")
)

// CreateContainer creates new container
func (p *Pool) CreateContainer(repository string, opts ...Option) (string, error) {
	if len(repository) == 0 {
		return "", ErrRepositoryEmpty
	}

	options := &Options{
		Tag: "latest",
	}

	for _, opt := range opts {
		opt(options)
	}

	exposedPort := nat.PortSet{}
	if len(options.ExposedPorts) > 0 {
		for _, port := range options.ExposedPorts {
			exposedPort[nat.Port(port)] = struct{}{}
		}
	}

	var mounts []mount.Mount
	if len(options.Mounts) > 0 {
		mounts = make([]mount.Mount, 0, len(options.Mounts))
		for _, target := range options.Mounts {
			source := path.Join(getCurrentDir(), "mount", target)
			if err := os.MkdirAll(source, 0777); err != nil {
				return "", err
			}
			mounts = append(mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: source,
				Target: target,
			})
		}
	}

	if _, _, err := p.client.ImageInspectWithRaw(p.ctx, fmt.Sprintf("%v:%v", repository, options.Tag)); err != nil {
		out, err := p.client.ImagePull(p.ctx, fmt.Sprintf("%v:%v", repository, options.Tag), types.ImagePullOptions{})
		if err != nil {
			return "", fmt.Errorf("ImagePull: %w", err)
		}
		io.Copy(os.Stdout, out)
	}

	resp, err := p.client.ContainerCreate(p.ctx,
		&containertypes.Config{
			Image:        fmt.Sprintf("%v:%v", repository, options.Tag),
			Env:          options.Env,
			Cmd:          options.Cmd,
			ExposedPorts: exposedPort,
		},
		&containertypes.HostConfig{
			RestartPolicy:   containertypes.RestartPolicy{Name: "always"},
			PublishAllPorts: true,
			Mounts:          mounts,
		}, nil, nil, options.Name)
	if err != nil {
		return "", fmt.Errorf("ContainerCreate: %w", err)
	}

	if err := p.client.ContainerStart(p.ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", fmt.Errorf("ContainerStart: %w", err)
	}

	info, err := p.client.ContainerInspect(p.ctx, resp.ID)
	if err != nil {
		return "", fmt.Errorf("ContainerInspect: %w", err)
	}

	p.containers[resp.ID] = info

	return resp.ID, nil
}

// GetContainerInfo get container infomation by id
func (p *Pool) GetContainerInfo(id string) (types.ContainerJSON, error) {
	containerInfo, ok := p.containers[id]
	if !ok {
		return types.ContainerJSON{}, ErrContainerNotFound
	}

	return containerInfo, nil
}

// GetHostPort get host port for exposed docker port
func (p *Pool) GetHostPort(containerID, dockerPort string) (string, error) {
	containerInfo, err := p.GetContainerInfo(containerID)
	if err != nil {
		return "", fmt.Errorf("GetContainerInfo: %w", err)
	}

	portMap := containerInfo.NetworkSettings.NetworkSettingsBase.Ports

	for port, portBinding := range portMap {
		if strings.Contains(string(port), dockerPort) {
			return fmt.Sprintf("localhost:%v", portBinding[0].HostPort), nil
		}
	}

	return "", errors.New("port not found")
}

// removeContainers removes created containers.
func (p *Pool) removeContainers() error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultDockerActionTimeout)
	defer cancel()

	for id, _ := range p.containers {
		if err := p.client.ContainerRemove(ctx, id, types.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		}); err != nil {
			return fmt.Errorf("ContainerRemove: %w", err)
		}
	}

	return nil
}
