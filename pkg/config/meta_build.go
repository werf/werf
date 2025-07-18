package config

type MetaBuild struct {
	CacheVersion string
	Platform     []string
	Staged       bool
	ImageSpec    *ImageSpec
}
