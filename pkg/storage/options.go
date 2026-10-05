package storage

import "github.com/werf/werf/v2/pkg/docker_registry"

type Options struct {
	withCache             bool
	dockerRegistryOptions []docker_registry.Option
}

func makeOptions(opts ...Option) Options {
	opt := Options{}
	for _, o := range opts {
		o(&opt)
	}

	return opt
}

type Option func(*Options)

func WithCache() Option {
	return func(o *Options) {
		o.withCache = true
		o.dockerRegistryOptions = append(o.dockerRegistryOptions, docker_registry.WithCachedTags())
	}
}

// WithFreshListing requires a listing started after this request, even if another is in flight.
func WithFreshListing() Option {
	return func(o *Options) {
		o.withCache = false
		o.dockerRegistryOptions = append(o.dockerRegistryOptions, docker_registry.WithFreshTags())
	}
}
