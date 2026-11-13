package config

import "github.com/werf/werf/v3/pkg/image"

type MetaBuild struct {
	CacheVersion string
	Platform     []string
	Staged       bool
	ImageSpec    *ImageSpec
	Compression  image.LayerCompression
}
