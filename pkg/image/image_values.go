package image

import (
	"fmt"
	"path"

	"github.com/google/go-containerregistry/pkg/name"
)

func BuildImageValuesMap(infoGetter *InfoGetter) (map[string]interface{}, error) {
	tag, err := name.NewTag(infoGetter.GetName())
	if err != nil {
		return nil, fmt.Errorf("build image values map: %w", err)
	}

	return buildValuesMap(
		tag.RegistryStr(),
		tag.RepositoryStr(),
		path.Base(tag.RepositoryStr()),
		tag.TagStr(),
		infoGetter.Digest,
		tag.Context().Name(),
		infoGetter.GetName(),
	), nil
}

func RebuildImageValuesMap(newRef, digest string) (map[string]interface{}, error) {
	tag, err := name.NewTag(newRef)
	if err != nil {
		return nil, fmt.Errorf("rebuild image values map: %w", err)
	}

	return buildValuesMap(
		tag.RegistryStr(),
		tag.RepositoryStr(),
		path.Base(tag.RepositoryStr()),
		tag.TagStr(),
		digest,
		tag.Context().Name(),
		newRef,
	), nil
}

func BuildStubImageValuesMap(repo, tag string) map[string]interface{} {
	return buildValuesMap(
		"REGISTRY",
		"NAMESPACE/NAME",
		"NAME",
		tag,
		"DIGEST",
		repo,
		fmt.Sprintf("%s:%s", repo, tag),
	)
}

func buildValuesMap(registry, repository, imageName, tag, digest, image, refTag string) map[string]interface{} {
	namespace := path.Dir(repository)

	return map[string]interface{}{
		"registry":   registry,
		"namespace":  namespace,
		"name":       imageName,
		"tag":        tag,
		"digest":     digest,
		"tag_digest": fmt.Sprintf("%s@%s", tag, digest),

		"image":      image,
		"repository": repository,

		"ref":     fmt.Sprintf("%s@%s", refTag, digest),
		"ref_tag": refTag,

		"repository_ref": fmt.Sprintf("%s:%s@%s", repository, tag, digest),
		"repository_tag": fmt.Sprintf("%s:%s", repository, tag),

		"name_ref": fmt.Sprintf("%s:%s@%s", imageName, tag, digest),
		"name_tag": fmt.Sprintf("%s:%s", imageName, tag),
	}
}
