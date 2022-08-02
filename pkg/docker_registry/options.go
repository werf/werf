package docker_registry

const OptionCachedTagsDefault = false

type Options struct {
	cachedTags bool
}

func makeOptions(opts ...Option) Options {
	opt := Options{
		cachedTags: OptionCachedTagsDefault,
	}
	for _, o := range opts {
		o(&opt)
	}

	return opt
}

type Option func(*Options)

func WithCachedTags() Option {
	return func(o *Options) {
		o.cachedTags = true
	}
}
