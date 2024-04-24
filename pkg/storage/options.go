package storage

import "github.com/werf/werf/v2/pkg/docker_registry"

type Options struct {
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
		o.dockerRegistryOptions = append(o.dockerRegistryOptions, docker_registry.WithCachedTags())
	}
}
