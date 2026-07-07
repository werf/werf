package common

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/storage"
)

type ResolveReposOptions struct {
	ImagesRepoRequired bool
	MetaRepoRequired   bool
}

func ResolveRepos(ctx context.Context, cmdData *CmdData, opts ResolveReposOptions) error {
	repoSet := cmdData.Repo != nil && *cmdData.Repo.Address != ""

	cacheFrom := GetCacheFrom(cmdData)
	cacheTo := GetCacheTo(cmdData)
	imagesRepo := GetImagesRepo(cmdData)
	metaRepo := ""
	if cmdData.MetaRepo != nil {
		metaRepo = *cmdData.MetaRepo
	}
	secondary := GetSecondaryStagesStorage(cmdData)
	cacheRepo := GetCacheStagesStorage(cmdData)

	granularSet := len(cacheFrom) > 0 || len(cacheTo) > 0 || metaRepo != ""

	if repoSet && granularSet {
		return fmt.Errorf("--repo is mutually exclusive with --cache-from, --cache-to and --meta-repo; use either --repo as a preset or the granular flags, not both (--images-repo may be combined with --repo to override the images destination)")
	}

	if repoSet && imagesRepo == "" {
		imagesRepo = *cmdData.Repo.Address
		*cmdData.ImagesRepo = imagesRepo
	}

	if len(secondary) > 0 {
		logboek.Context(ctx).Warn().LogF("DEPRECATED: --secondary-repo is deprecated, use --cache-from instead\n")
		cacheFrom = append(append([]string{}, cacheFrom...), secondary...)
		*cmdData.CacheFrom = cacheFrom
		*cmdData.SecondaryStagesStorage = nil
	}

	// werf 2.x --cache-repo was both a read and a write cache, so fold it into
	// both lists. The fold runs after the mutual-exclusion check on purpose:
	// --repo + --cache-repo was the canonical 2.x usage and must keep working.
	if len(cacheRepo) > 0 {
		logboek.Context(ctx).Warn().LogF("DEPRECATED: --cache-repo ($WERF_CACHE_REPO_*) is deprecated, use --cache-from and --cache-to instead\n")
		cacheFrom = append(append([]string{}, cacheFrom...), cacheRepo...)
		*cmdData.CacheFrom = cacheFrom
		*cmdData.CacheTo = append(append([]string{}, cacheTo...), cacheRepo...)
		if cmdData.CacheStagesStorage != nil {
			*cmdData.CacheStagesStorage = nil
		}
	}

	if opts.ImagesRepoRequired && !repoSet && imagesRepo == "" {
		return fmt.Errorf("--images-repo=ADDRESS (or --repo=ADDRESS) is required")
	}
	if opts.MetaRepoRequired && !repoSet && metaRepo == "" {
		return fmt.Errorf("--meta-repo=ADDRESS (or --repo=ADDRESS) is required")
	}

	if !repoSet && len(cacheFrom) == 0 {
		*cmdData.CacheFrom = []string{storage.LocalStorageAddress}
	}

	return nil
}
