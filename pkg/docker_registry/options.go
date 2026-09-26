package docker_registry

import "time"

const OptionCachedTagsDefault = false

type Options struct {
	cachedTags bool
	tagsMaxAge time.Duration
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

// WithTagsMaxAge allows serving the tags listing from the cache when it was fetched from the
// registry no longer than maxAge ago, instead of requiring a fresh listing.
func WithTagsMaxAge(maxAge time.Duration) Option {
	return func(o *Options) {
		o.tagsMaxAge = maxAge
	}
}
