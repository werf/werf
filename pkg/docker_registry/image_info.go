package docker_registry

import (
	v1 "github.com/google/go-containerregistry/pkg/v1"

	"github.com/werf/werf/pkg/image"
)

func NewImageInfoFromRegistryConfig(ref string, cfg *v1.ConfigFile) *image.Info {
	repository, tag := image.ParseRepositoryAndTag(ref)
	return &image.Info{
		Name:              ref,
		Repository:        repository,
		Tag:               tag,
		Labels:            cfg.Config.Labels,
		OnBuild:           cfg.Config.OnBuild,
		Env:               cfg.Config.Env,
		CreatedAtUnixNano: cfg.Created.UnixNano(),
		RepoDigest:        "", // TODO
		ID:                "", // TODO
		ParentID:          "", // TODO
		Size:              0,  // TODO
	}
}
