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

	parentID := inspect.Config.Image
	if parentID == "" {
		if id, ok := inspect.Config.Labels[image.WerfBaseImageIDLabel]; ok { // built with werf and buildah backend
			parentID = id
		}
	}

	var created string
	if inspect.Created != "" {
		created = inspect.Created
	} else {
		created = "1970-01-01T00:00:00Z"
	}

	return &image.Info{
		Name:              ref,
		Repository:        repository,
		Tag:               tag,
		Labels:            inspect.Config.Labels,
		OnBuild:           inspect.Config.OnBuild,
		Env:               inspect.Config.Env,
		CreatedAtUnixNano: image.MustParseTimestampString(created).UnixNano(),
		RepoDigest:        repoDigest,
		ID:                inspect.ID,
		ParentID:          parentID,
		Size:              inspect.Size,
	}
}
