package easydocker

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const (
	defaultDockerActionTimeout = 5 * time.Second
	defaultMaxWaitRetryTime    = time.Minute
)

// Pool represents a connection to the docker API and is used to create and remove container
type Pool struct {
	ctx    context.Context
	cancel context.CancelFunc

	client          client.APIClient // used to interact with docker API
	maxWaitForRetry time.Duration

	networks   map[string]types.NetworkResource // network id => network info
	networksMu sync.RWMutex                     // used to guard networks map

	containers   map[string]types.ContainerJSON // container id => container info
	containersMu sync.RWMutex                   // used to guard containers map
}

// NewPool creates a new pool.
func NewPool(endpoint string) (*Pool, error) {
	if len(endpoint) == 0 {
		if len(os.Getenv("DOCKER_HOST")) != 0 {
			endpoint = os.Getenv("DOCKER_HOST")
		} else if len(os.Getenv("DOCKER_URL")) != 0 {
			endpoint = os.Getenv("DOCKER_URL")
		} else if runtime.GOOS == "windows" {
			endpoint = "http://localhost:2375"
		} else {
			endpoint = "unix:///var/run/docker.sock"
		}
	}

	cli, err := client.NewClientWithOpts(client.WithHost(endpoint), client.WithTimeout(5*time.Minute), client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("NewClienWithOpts: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Pool{
		ctx:    ctx,
		cancel: cancel,

		client:          cli,
		maxWaitForRetry: defaultMaxWaitRetryTime,

		networks: make(map[string]types.NetworkResource),
	}, nil
}

// Close stops running actions, removes created networks and containers.
func (p *Pool) Close() error {
	p.cancel()
	if err := p.removeContainers(); err != nil {
		return err
	}
	if err := p.removeNetworks(); err != nil {
		return err
	}
	return nil
}
