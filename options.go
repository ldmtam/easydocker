package easydocker

// Options is used to pass when running a container.
type Options struct {
	Name         string
	Tag          string
	Env          []string
	Cmd          []string
	NetworkID    string
	ExposedPorts []string
	Mounts       []string
}

// Option wraps function that modifies Options
type Option func(*Options)

// WithName assign name for container
func WithName(name string) Option {
	return func(opts *Options) {
		opts.Name = name
	}
}

// WithTag repository tag
func WithTag(tag string) Option {
	return func(opts *Options) {
		opts.Tag = tag
	}
}

// WithEnvironment add environment variables for container
func WithEnvironment(env ...string) Option {
	return func(opts *Options) {
		opts.Env = append(opts.Env, env...)
	}
}

// WithCmd add command for container
func WithCmd(cmd ...string) Option {
	return func(opts *Options) {
		opts.Cmd = append(opts.Cmd, cmd...)
	}
}

// WithNetwork assign docker network for container
func WithNetwork(networkID string) Option {
	return func(opts *Options) {
		opts.NetworkID = networkID
	}
}

// WithExposedPorts maps container port to host port
func WithExposedPorts(exposedPorts ...string) Option {
	return func(opts *Options) {
		opts.ExposedPorts = append(opts.ExposedPorts, exposedPorts...)
	}
}

// WithMounts mounts folder in container to host
func WithMounts(mounts ...string) Option {
	return func(opts *Options) {
		opts.Mounts = append(opts.Mounts, mounts...)
	}
}
