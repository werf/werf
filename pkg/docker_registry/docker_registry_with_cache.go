package docker_registry

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/werf/logboek"
	registry_api "github.com/werf/werf/v2/pkg/docker_registry/api"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/opstats"
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
	cachedTagsMap   *sync.Map
	cacheWriteMu    sync.Mutex
	listingSequence uint64
	storedSequence  map[string]uint64
	publishedTags   map[string]map[string]uint64

	listTagsQueryGroup *singleflight.Group
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
	outcome, sharedLookup := opstats.CacheOutcomeBypass, false
	if makeOptions(opts...).cachedTags {
		outcome = opstats.CacheOutcomeMiss
	}
	defer func() {
		opstats.CountCacheLookup(ctx, opstats.OperationRegistryTagsList, opstats.CacheLayerMemory, outcome, sharedLookup)
	}()
	cachedTagsID := r.mustGetCachedTagsID(reference)
	if makeOptions(opts...).freshTags {
		outcome = opstats.CacheOutcomeBypass
		sequence := r.nextListingSequence()
		tags, err := r.Interface.Tags(ctx, reference, opts...)
		if err != nil {
			return nil, fmt.Errorf("fetch fresh tags for repo %q: %w", reference, err)
		}
		return r.storeTagsToCache(cachedTagsID, tags, sequence), nil
	}
	if tags, ok := r.tryLoadTagsFromCache(cachedTagsID, opts...); ok {
		outcome = opstats.CacheOutcomeHit
		opstats.CountEvent(ctx, opstats.EventRegistryTagsCacheHit)
		return tags, nil
	}

	// Use singleflight to avoid multiple concurrent calls to the registry for the same reference
	// This is useful when multiple goroutines try to fetch tags for the same reference at the same time.
	// Will perform only one call to the registry and share the result among all goroutines.
	leader := false
	newTagsResp, err, shared := r.listTagsQueryGroup.Do(cachedTagsID, func() (interface{}, error) {
		leader = true
		sequence := r.nextListingSequence()
		tags, err := r.Interface.Tags(ctx, reference, opts...)
		if err == nil {
			tags = r.storeTagsToCache(cachedTagsID, tags, sequence)
		}
		return tags, err
	})

	sharedLookup = !leader
	if shared {
		opstats.CountEvent(ctx, opstats.EventRegistryTagsSharedResult)
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

func (r *DockerRegistryWithCache) nextListingSequence() uint64 {
	r.cacheWriteMu.Lock()
	defer r.cacheWriteMu.Unlock()
	r.listingSequence++
	return r.listingSequence
}

// A slow earlier listing must not replace a newer authoritative result in the cache.
func (r *DockerRegistryWithCache) storeTagsToCache(id string, tags []string, sequence uint64) []string {
	r.cacheWriteMu.Lock()
	defer r.cacheWriteMu.Unlock()
	if r.storedSequence[id] > sequence {
		if value, ok := r.cachedTagsMap.Load(id); ok {
			if cached, err := castTagsList(value); err == nil {
				return cached
			}
		}
		return tags
	}
	if r.storedSequence == nil {
		r.storedSequence = make(map[string]uint64)
	}
	r.storedSequence[id] = sequence
	pending := make(map[string]uint64)
	for tag, publishedAt := range r.publishedTags[id] {
		if publishedAt >= sequence && !slices.Contains(tags, tag) {
			tags = append(slices.Clone(tags), tag)
			pending[tag] = publishedAt
		}
	}
	if r.publishedTags != nil {
		r.publishedTags[id] = pending
	}
	r.cachedTagsMap.Store(id, tags)
	return tags
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
	if err := r.Interface.TagRepoImage(ctx, repoImage, tag); err != nil {
		return err
	}
	AddCachedTag(ctx, r, repoImage.Repository+":"+tag)
	return nil
}

func (r *DockerRegistryWithCache) PushImage(ctx context.Context, reference string, opts *PushImageOptions) error {
	if err := r.Interface.PushImage(ctx, reference, opts); err != nil {
		return err
	}
	AddCachedTag(ctx, r, reference)
	return nil
}

func (r *DockerRegistryWithCache) MutateAndPushImage(ctx context.Context, sourceReference, destinationReference string, opts ...registry_api.MutateOption) error {
	if err := r.Interface.MutateAndPushImage(ctx, sourceReference, destinationReference, opts...); err != nil {
		return err
	}
	AddCachedTag(ctx, r, destinationReference)
	return nil
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

							_, err := r.Tags(ctxWithTimeout, repo)
							if err != nil {
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

func (r *DockerRegistryWithCache) TagsByPrefix(ctx context.Context, reference, prefix string) ([]string, error) {
	opstats.CountCacheLookup(ctx, opstats.Operation("registry: tags list by prefix"), opstats.CacheLayerMemory, opstats.CacheOutcomeBypass, false)
	return r.Interface.TagsByPrefix(ctx, reference, prefix)
}

func (r *DockerRegistryWithCache) ManifestExists(ctx context.Context, reference string) (bool, error) {
	return r.Interface.ManifestExists(ctx, reference)
}

// AddCachedTag preserves observed tags across older listings without creating a full inventory.
func AddCachedTag(ctx context.Context, registry Interface, reference string) {
	r, ok := registry.(*DockerRegistryWithCache)
	if !ok {
		return
	}
	parts, err := r.parseReferenceParts(reference)
	if err != nil {
		logboek.Context(ctx).Debug().LogF("Parse published reference for tags cache: %s\n", err)
		return
	}
	if parts.tag == "" || parts.digest != "" {
		return
	}
	id := parts.registry + "/" + parts.repository
	r.cacheWriteMu.Lock()
	defer r.cacheWriteMu.Unlock()
	if r.publishedTags == nil {
		r.publishedTags = make(map[string]map[string]uint64)
	}
	if r.publishedTags[id] == nil {
		r.publishedTags[id] = make(map[string]uint64)
	}
	r.publishedTags[id][parts.tag] = r.listingSequence
	value, ok := r.cachedTagsMap.Load(id)
	if !ok {
		return
	}
	tags, err := castTagsList(value)
	if err != nil {
		return
	}
	if !slices.Contains(tags, parts.tag) {
		tags = append(slices.Clone(tags), parts.tag)
	}
	r.cachedTagsMap.Store(id, tags)
}

func (r *DockerRegistryWithCache) CopyImage(ctx context.Context, sourceReference, destinationReference string, opts CopyImageOptions) error {
	if err := r.Interface.CopyImage(ctx, sourceReference, destinationReference, opts); err != nil {
		return err
	}
	AddCachedTag(ctx, r, destinationReference)
	return nil
}

func (r *DockerRegistryWithCache) PushManifestList(ctx context.Context, reference string, opts ManifestListOptions) error {
	if err := r.Interface.PushManifestList(ctx, reference, opts); err != nil {
		return err
	}
	AddCachedTag(ctx, r, reference)
	return nil
}
