package docker_registry

import (
	"context"
	"fmt"
	"strings"
	"sync"

	v1 "github.com/google/go-containerregistry/pkg/v1"

	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/util"
)

type DockerRegistryWithCache struct {
	Interface
	cachedTagsMap      *sync.Map
	cachedTagsMutexMap *sync.Map
}

func newDockerRegistryWithCache(dockerRegistry Interface) *DockerRegistryWithCache {
	return &DockerRegistryWithCache{
		Interface:          dockerRegistry,
		cachedTagsMap:      &sync.Map{},
		cachedTagsMutexMap: &sync.Map{},
	}
}

func (r *DockerRegistryWithCache) Tags(ctx context.Context, reference string, opts ...Option) ([]string, error) {
	o := makeOptions(opts...)
	return r.withCachedTags(reference, func(cachedTags []string, isExist bool) ([]string, error) {
		if isExist && o.cachedTags {
			return cachedTags, nil
		}

		return r.Interface.Tags(ctx, reference, opts...)
	})
}

func (r *DockerRegistryWithCache) IsTagExist(ctx context.Context, reference string, opts ...Option) (bool, error) {
	referenceParts, err := r.parseReferenceParts(reference)
	if err != nil {
		return false, err
	}

	referenceTag := referenceParts.tag
	if referenceTag == "" {
		panic(fmt.Sprintf("unexpected reference %q: tag required", reference))
	}

	repositoryAddress := strings.Join([]string{referenceParts.registry, referenceParts.repository}, "/")
	tags, err := r.Tags(ctx, repositoryAddress, opts...)
	if err != nil {
		return false, err
	}

	for _, tag := range tags {
		if referenceTag == tag {
			return true, nil
		}
	}

	return false, nil
}

func (r *DockerRegistryWithCache) TagRepoImage(ctx context.Context, repoImage *image.Info, tag string) error {
	defer r.mustAddTagToCachedTags(repoImage.Name)
	return r.Interface.TagRepoImage(ctx, repoImage, tag)
}

func (r *DockerRegistryWithCache) PushImage(ctx context.Context, reference string, opts *PushImageOptions) error {
	defer r.mustAddTagToCachedTags(reference)
	return r.Interface.PushImage(ctx, reference, opts)
}

func (r *DockerRegistryWithCache) MutateAndPushImage(ctx context.Context, sourceReference, destinationReference string, mutateConfigFunc func(v1.Config) (v1.Config, error)) error {
	defer r.mustAddTagToCachedTags(destinationReference)
	return r.Interface.MutateAndPushImage(ctx, sourceReference, destinationReference, mutateConfigFunc)
}

func (r *DockerRegistryWithCache) DeleteRepoImage(ctx context.Context, repoImage *image.Info) error {
	defer r.mustDeleteTagFromCachedTags(repoImage.Name)
	return r.Interface.DeleteRepoImage(ctx, repoImage)
}

func (r *DockerRegistryWithCache) mustAddTagToCachedTags(reference string) {
	_, err := r.withCachedTags(reference, func(tags []string, isExist bool) ([]string, error) {
		referenceParts, err := r.parseReferenceParts(reference)
		if err != nil {
			return nil, fmt.Errorf("unable to parse reference parts %q: %w", reference, err)
		}

		if !isExist {
			return nil, nil
		}

		tags = append(tags, referenceParts.tag)
		return tags, nil
	})
	if err != nil {
		panic(fmt.Sprintf("unexpected err: %s", err))
	}
}

func (r *DockerRegistryWithCache) mustDeleteTagFromCachedTags(reference string) {
	_, err := r.withCachedTags(reference, func(tags []string, isExist bool) ([]string, error) {
		referenceParts, err := r.parseReferenceParts(reference)
		if err != nil {
			return nil, fmt.Errorf("unable to parse reference parts %q: %w", reference, err)
		}

		if !isExist {
			return nil, nil
		}

		tags = util.ExcludeFromStringArray(tags, referenceParts.tag)
		return tags, nil
	})
	if err != nil {
		panic(fmt.Sprintf("unexpected err: %s", err))
	}
}

func (r *DockerRegistryWithCache) withCachedTags(reference string, f func([]string, bool) ([]string, error)) ([]string, error) {
	cachedTagsID := r.mustGetCachedTagsID(reference)

	mutex := util.MapLoadOrCreateMutex(r.cachedTagsMutexMap, cachedTagsID)
	mutex.Lock()
	defer mutex.Unlock()

	value, isExist := r.cachedTagsMap.Load(cachedTagsID)
	var tags []string
	if isExist {
		tags = value.([]string)
	}

	newTags, err := f(tags, isExist)
	if err != nil {
		return nil, err
	}

	r.cachedTagsMap.Store(cachedTagsID, newTags)
	return newTags, nil
}

func (r *DockerRegistryWithCache) mustGetCachedTagsID(reference string) string {
	referenceParts, err := r.parseReferenceParts(reference)
	if err != nil {
		panic(fmt.Sprintf("unexpected reference %q: %s", reference, err))
	}

	repositoryAddress := strings.Join([]string{referenceParts.registry, referenceParts.repository}, "/")
	return repositoryAddress
}
