package docker_registry

import (
	"context"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/werf/logboek"
	registry_api "github.com/werf/werf/v3/pkg/docker_registry/api"
	"github.com/werf/werf/v3/pkg/image"
)

const (
	defaultUpdaterPollInterval = 5 * time.Minute
	defaultUpdaterTaskTimeout  = 1 * time.Minute
)

const (
	updaterPollIntervalEnv = "WERF_PUBLISH_TAG_CACHE_POLL_INTERVAL"
	updaterTaskTimeoutEnv  = "WERF_PUBLISH_TAG_CACHE_TASK_TIMEOUT"
)

type DockerRegistryWithCache struct {
	Interface
	cachedTagsMap *sync.Map
	// cacheWriteMu serializes the load-modify-store sequences on cachedTagsMap. Entries stay
	// immutable, so readers need no lock.
	cacheWriteMu       sync.Mutex
	listTagsQueryGroup *singleflight.Group
}

// AddCachedTag records a tag published to the registry outside the registry client, so that the
// next cached tags lookup sees it. It only updates an already cached listing, never creates one,
// and never refreshes its freshness timestamp. A reference that is not a tagged one is ignored.
func AddCachedTag(ctx context.Context, registry Interface, reference string) {
	r, ok := registry.(*DockerRegistryWithCache)
	if !ok {
		return
	}

	referenceParts, err := r.parseReferenceParts(reference)
	if err != nil {
		logboek.Context(ctx).Debug().LogF("Not adding published reference %q to the tags cache: %s\n", reference, err)
		return
	}
	if referenceParts.tag == "" || referenceParts.digest != "" {
		logboek.Context(ctx).Debug().LogF("Not adding published reference %q to the tags cache: not a tagged reference\n", reference)
		return
	}
	cachedTagsID := strings.Join([]string{referenceParts.registry, referenceParts.repository}, "/")

	r.cacheWriteMu.Lock()
	defer r.cacheWriteMu.Unlock()

	value, ok := r.cachedTagsMap.Load(cachedTagsID)
	if !ok {
		return
	}

	entry, err := castTagsEntry(value)
	if err != nil {
		return
	}

	tags := entry.tags
	if !slices.Contains(tags, referenceParts.tag) {
		tags = append(slices.Clone(tags), referenceParts.tag)
	}

	pushedTags := maps.Clone(entry.pushedTags)
	if pushedTags == nil {
		pushedTags = map[string]time.Time{}
	}
	// Recorded even when the tag is already listed: a listing in flight may have snapshotted the
	// repo without it.
	pushedTags[referenceParts.tag] = time.Now()

	// updatedAt is left untouched: the listing itself is no fresher than it was.
	r.cachedTagsMap.Store(cachedTagsID, tagsCacheEntry{tags: tags, updatedAt: entry.updatedAt, pushedTags: pushedTags})

	logboek.Context(ctx).Debug().LogF("Added published tag %q to the tags cache of %q\n", referenceParts.tag, cachedTagsID)
}

func newDockerRegistryWithCache(ctx context.Context, dockerRegistry Interface) *DockerRegistryWithCache {
	r := &DockerRegistryWithCache{
		Interface:          dockerRegistry,
		cachedTagsMap:      &sync.Map{},
		listTagsQueryGroup: &singleflight.Group{},
	}

	if os.Getenv("WERF_DISABLE_PUBLISH_TAG_CACHE_SYNC") == "1" {
		pollInterval := readDurationEnv(ctx, updaterPollIntervalEnv, defaultUpdaterPollInterval)
		taskTimeout := readDurationEnv(ctx, updaterTaskTimeoutEnv, defaultUpdaterTaskTimeout)
		r.startBackgroundCacheUpdater(ctx, pollInterval, taskTimeout)
	}

	return r
}

func (r *DockerRegistryWithCache) Tags(ctx context.Context, reference string, opts ...Option) ([]string, error) {
	return r.getTagsListFromRegistry(ctx, reference, opts...)
}

func (r *DockerRegistryWithCache) tryLoadTagsFromCache(cachedTagsID string, opts ...Option) ([]string, bool) {
	o := makeOptions(opts...)
	value, ok := r.cachedTagsMap.Load(cachedTagsID)
	if !ok {
		return nil, false
	}

	entry, err := castTagsEntry(value)
	if err != nil {
		return nil, false
	}

	if o.cachedTags {
		return entry.tags, true
	}
	if o.tagsMaxAge > 0 && time.Since(entry.updatedAt) <= o.tagsMaxAge {
		return entry.tags, true
	}

	return nil, false
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
		startedAt := time.Now()
		tags, err := r.Interface.Tags(ctx, reference, opts...)
		if err != nil {
			return nil, err
		}
		// Storing inside the singleflight call keeps a listing from dropping tags published while
		// it was in flight, for the cache and for every waiter alike.
		return r.storeTagsToCache(cachedTagsID, tags, startedAt), nil
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

	return newTagsList, nil
}

func (r *DockerRegistryWithCache) storeTagsToCache(cachedTagsID string, tags []string, listingStartedAt time.Time) []string {
	r.cacheWriteMu.Lock()
	defer r.cacheWriteMu.Unlock()

	pushedTags := map[string]time.Time{}
	if value, ok := r.cachedTagsMap.Load(cachedTagsID); ok {
		if entry, err := castTagsEntry(value); err == nil && len(entry.pushedTags) > 0 {
			cloned := false
			for tag, pushedAt := range entry.pushedTags {
				// A listing started after the publication is authoritative and drops the tag,
				// so an externally deleted tag cannot stay cached forever.
				if !pushedAt.After(listingStartedAt) || slices.Contains(tags, tag) {
					continue
				}
				if !cloned {
					tags, cloned = slices.Clone(tags), true
				}
				tags = append(tags, tag)
				pushedTags[tag] = pushedAt
			}
		}
	}

	r.cachedTagsMap.Store(cachedTagsID, tagsCacheEntry{tags: tags, updatedAt: time.Now(), pushedTags: pushedTags})
	return tags
}

type tagsCacheEntry struct {
	tags      []string
	updatedAt time.Time
	// pushedTags holds tags published locally but not yet confirmed by a registry listing.
	pushedTags map[string]time.Time
}

func castTagsList(tagsList interface{}) ([]string, error) {
	switch v := tagsList.(type) {
	case []string:
		return v, nil
	default:
		return nil, fmt.Errorf("unexpected type %T for tags", v)
	}
}

func castTagsEntry(value interface{}) (tagsCacheEntry, error) {
	switch v := value.(type) {
	case tagsCacheEntry:
		return v, nil
	default:
		return tagsCacheEntry{}, fmt.Errorf("unexpected type %T for tags cache entry", v)
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

							if _, err := r.Tags(ctxWithTimeout, repo); err != nil {
								logboek.Context(ctx).Debug().LogF("Failed to update tag cache for %q: %s\n", repo, err)
								return err
							}

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

func readDurationEnv(ctx context.Context, envName string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(envName))
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		logboek.Context(ctx).Warn().LogF("WARNING: Invalid %s=%q, using default %s: %s\n", envName, value, fallback, err)
		return fallback
	}

	if parsed <= 0 {
		logboek.Context(ctx).Warn().LogF("WARNING: Non-positive %s=%q, using default %s\n", envName, value, fallback)
		return fallback
	}

	return parsed
}
