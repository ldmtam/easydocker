package easydocker

import (
	"context"
	"errors"
	"fmt"

	"github.com/docker/docker/api/types"
)

var (
	// ErrContainerNotFound container not found
	ErrContainerNotFound = errors.New("container not found")
)

// GetContainer get container infomation by id
func (p *Pool) GetContainer(id string) (types.ContainerJSON, error) {
	p.containersMu.RLock()
	defer p.containersMu.RUnlock()

	containerInfo, ok := p.containers[id]
	if !ok {
		return types.ContainerJSON{}, ErrContainerNotFound
	}

	return containerInfo, nil
}

// removeContainers removes created containers.
func (p *Pool) removeContainers() error {
	p.containersMu.Lock()
	defer p.containersMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), defaultDockerActionTimeout)
	defer cancel()

	for id, _ := range p.containers {
		if err := p.client.ContainerRemove(ctx, id, types.ContainerRemoveOptions{
			RemoveVolumes: true,
			RemoveLinks:   true,
			Force:         true,
		}); err != nil {
			return fmt.Errorf("ContainerRemove: %w", err)
		}
	}

	return nil
}
