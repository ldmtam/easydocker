package easydocker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
)

// CreateNetwork creates docker network
func (p *Pool) CreateNetwork(name string) (types.NetworkResource, error) {
	resp, err := p.client.NetworkCreate(p.ctx, name, types.NetworkCreate{
		CheckDuplicate: true,
	})
	if err != nil {
		return types.NetworkResource{}, fmt.Errorf("NetworkCreate: %w", err)
	}

	networkResource, err := p.client.NetworkInspect(p.ctx, resp.ID, types.NetworkInspectOptions{})
	if err != nil {
		return types.NetworkResource{}, fmt.Errorf("NetworkInspect: %w", err)
	}

	p.networksMu.Lock()
	defer p.networksMu.Unlock()
	p.networks[networkResource.ID] = networkResource

	return networkResource, nil
}

// removeNetworks removes created networks
func (p *Pool) removeNetworks() error {
	p.networksMu.Lock()
	defer p.networksMu.Unlock()

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
