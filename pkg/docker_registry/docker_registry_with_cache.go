package docker_registry

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/werf/logboek"
	registry_api "github.com/werf/werf/v2/pkg/docker_registry/api"
	"github.com/werf/werf/v2/pkg/image"
)

const (
	defaultUpdaterPollInterval = 5 * time.Minute
	defaultUpdaterTaskTimeout  = 1 * time.Minute
)

type DockerRegistryWithCache struct {
	Interface
	cachedTagsMap *sync.Map

	listTagsQueryGroup *singleflight.Group
}

func newDockerRegistryWithCache(ctx context.Context, dockerRegistry Interface) *DockerRegistryWithCache {
	r := &DockerRegistryWithCache{
		Interface:          dockerRegistry,
		cachedTagsMap:      &sync.Map{},
		listTagsQueryGroup: &singleflight.Group{},
	}

	if os.Getenv("WERF_DISABLE_PUBLISH_TAG_CACHE_SYNC") == "1" {
		r.startBackgroundCacheUpdater(ctx, defaultUpdaterPollInterval, defaultUpdaterTaskTimeout)
	}

	return r
}

func (r *DockerRegistryWithCache) Tags(ctx context.Context, reference string, opts ...Option) ([]string, error) {
	return r.getTagsListFromRegistry(ctx, reference, opts...)
}

func (r *DockerRegistryWithCache) tryLoadTagsFromCache(cachedTagsID string, opts ...Option) ([]string, bool) {
	o := makeOptions(opts...)
	value, ok := r.cachedTagsMap.Load(cachedTagsID)
	if !ok || !o.cachedTags {
		return nil, false
	}

	tagsList, err := castTagsList(value)
	if err != nil {
		return nil, false
	}

	return tagsList, true
}

func (r *DockerRegistryWithCache) getTagsListFromRegistry(ctx context.Context, reference string, opts ...Option) ([]string, error) {
	cachedTagsID := r.mustGetCachedTagsID(reference)
	if tags, ok := r.tryLoadTagsFromCache(cachedTagsID, opts...); ok {
		return tags, nil
	}

	// Use singleflight to avoid multiple concurrent calls to the registry for the same reference
	// This is useful when multiple goroutines try to fetch tags for the same reference at the same time.
	// Will perform only one call to the registry and share the result among all goroutines.
	newTagsResp, err, shared := r.listTagsQueryGroup.Do(cachedTagsID, func() (interface{}, error) {
		tags, err := r.Interface.Tags(ctx, reference, opts...)
		return tags, err
	})

	if shared {
		logboek.Context(ctx).Debug().LogF("Query list tags for %q was reused\n", cachedTagsID)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to fetch tags for repo %q: %w", reference, err)
	}

	newTagsList, err := castTagsList(newTagsResp)
	if err != nil {
		return nil, err
	}

	r.cachedTagsMap.Store(cachedTagsID, newTagsList)
	return newTagsList, nil
}

func castTagsList(tagsList interface{}) ([]string, error) {
	switch v := tagsList.(type) {
	case []string:
		return v, nil
	default:
		return nil, fmt.Errorf("unexpected type %T for tags", v)
	}
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
	tags, err := Tags(ctx, r, repositoryAddress, opts...)
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
	return r.Interface.TagRepoImage(ctx, repoImage, tag)
}

func (r *DockerRegistryWithCache) PushImage(ctx context.Context, reference string, opts *PushImageOptions) error {
	return r.Interface.PushImage(ctx, reference, opts)
}

func (r *DockerRegistryWithCache) MutateAndPushImage(ctx context.Context, sourceReference, destinationReference string, opts ...registry_api.MutateOption) error {
	return r.Interface.MutateAndPushImage(ctx, sourceReference, destinationReference, opts...)
}

func (r *DockerRegistryWithCache) DeleteRepoImage(ctx context.Context, repoImage *image.Info) error {
	return r.Interface.DeleteRepoImage(ctx, repoImage)
}

func (r *DockerRegistryWithCache) mustGetCachedTagsID(reference string) string {
	referenceParts, err := r.parseReferenceParts(reference)
	if err != nil {
		panic(fmt.Sprintf("unexpected reference %q: %s", reference, err))
	}

	repositoryAddress := strings.Join([]string{referenceParts.registry, referenceParts.repository}, "/")
	return repositoryAddress
}

func (r *DockerRegistryWithCache) startBackgroundCacheUpdater(ctx context.Context, pollInterval, timeout time.Duration) {
	logboek.Context(ctx).Info().LogLn("Background docker registry cache updater started")
	go func() {
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.cachedTagsMap.Range(func(key, _ any) bool {
					cachedTagsID, ok := key.(string)
					if !ok {
						return true
					}
					go func(repo string) {
						if err := logboek.Context(ctx).Info().LogProcess("Update repo %s cache in background", repo).DoError(func() error {
							ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
							defer cancel()

							tags, err := r.Tags(ctxWithTimeout, repo)
							if err != nil {
								logboek.Context(ctx).Debug().LogF("Failed to update tag cache for %q: %s\n", repo, err)
								return err
							}

							r.cachedTagsMap.Store(repo, tags)
							logboek.Context(ctx).Debug().LogF("Updated tag cache for %q\n", repo)
							return nil
						}); err != nil {
							return
						}
					}(cachedTagsID)
					return true
				})
			}
		}
	}()
}
