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

	granularSet := len(cacheFrom) > 0 || len(cacheTo) > 0 || metaRepo != ""

	if repoSet && granularSet {
		return fmt.Errorf("--repo is mutually exclusive with the granular registry flags (--cache-from, --cache-to, --images-repo, --meta-repo); use either --repo as a preset or the granular flags, not both")
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
