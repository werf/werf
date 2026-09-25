package storage

import "github.com/werf/werf/v3/pkg/docker_registry"

type Options struct {
	dockerRegistryOptions []docker_registry.Option
	withCache             bool
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
