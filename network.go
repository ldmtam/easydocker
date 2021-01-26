package easydocker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
)

// CreateNetwork creates docker network
func (p *Pool) CreateNetwork(name string) (string, error) {
	resp, err := p.client.NetworkCreate(p.ctx, name, types.NetworkCreate{
		CheckDuplicate: true,
	})
	if err != nil {
		return "", fmt.Errorf("NetworkCreate: %w", err)
	}

	networkResource, err := p.client.NetworkInspect(p.ctx, resp.ID, types.NetworkInspectOptions{})
	if err != nil {
		return "", fmt.Errorf("NetworkInspect: %w", err)
	}

	p.networks[networkResource.ID] = networkResource

	return networkResource.ID, nil
}

// removeNetworks removes created networks
func (p *Pool) removeNetworks() error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultDockerActionTimeout)
	defer cancel()

	for id, resource := range p.networks {
		for container := range resource.Containers {
			if err := p.client.NetworkDisconnect(ctx, id, container, true); err != nil {
				return fmt.Errorf("NetworkDisconnect: %w", err)
			}
		}

		if err := p.client.NetworkRemove(ctx, id); err != nil {
			return fmt.Errorf("NetworkRemove: %w", err)
		}
	}

	return nil
}
