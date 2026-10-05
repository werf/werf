package docker_registry

const OptionCachedTagsDefault = false

type Options struct {
	cachedTags bool
	freshTags  bool
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

// WithFreshTags starts a new registry listing instead of joining an earlier request.
func WithFreshTags() Option {
	return func(o *Options) { o.freshTags = true }
}
