package docker

import (
	"strings"

	"github.com/docker/docker/api/types"

	"github.com/werf/werf/pkg/image"
)

func NewInfoFromInspect(ref string, inspect *types.ImageInspect) *image.Info {
	var repository, tag, repoDigest string
	if !strings.HasPrefix(ref, "sha256:") {
		repository, tag = image.ParseRepositoryAndTag(ref)
		repoDigest = image.ExtractRepoDigest(inspect.RepoDigests, repository)
	}

	var parentID string
	if id, ok := inspect.Config.Labels["werf.io/base-image-id"]; ok {
		parentID = id
	} else {
		// TODO(1.3): Legacy compatibility mode
		parentID = inspect.Config.Image
	}

	return &image.Info{
		Name:              ref,
		Repository:        repository,
		Tag:               tag,
		Labels:            inspect.Config.Labels,
		OnBuild:           inspect.Config.OnBuild,
		Env:               inspect.Config.Env,
		CreatedAtUnixNano: image.MustParseTimestampString(inspect.Created).UnixNano(),
		RepoDigest:        repoDigest,
		ID:                inspect.ID,
		ParentID:          parentID,
		Size:              inspect.Size,
	}
}
