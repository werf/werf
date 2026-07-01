package common

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/storage"
)

// ResolveReposOptions carries per-command requiredness hints for the granular
// registry model. --repo preset satisfies all of them.
type ResolveReposOptions struct {
	// ImagesRepoRequired: build --push / converge need at least one images destination.
	ImagesRepoRequired bool
	// MetaRepoRequired: cleanup needs a meta destination.
	MetaRepoRequired bool
}

// ResolveRepos validates and normalizes the registry-model flags in place on
// cmdData. It runs once per command before any storage is assembled.
//
// Rules (v3 registry rework):
//   - --repo and any granular flag (--cache-from/--cache-to/--images-repo/
//     --meta-repo) or deprecated alias (--secondary-repo/--final-repo) are
//     mutually exclusive: combining them is an error.
//   - --repo preset fans out to cache-from = cache-to = images-repo = meta-repo
//     = repo, preserving current behavior bit-for-bit (getters keep reading Repo).
//   - --secondary-repo is a deprecated alias for --cache-from; --final-repo for
//     --images-repo. Both emit a deprecation warning and feed the new lists.
//   - --cache-from defaults to [:local] when nothing is configured.
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
	finalRepo := ""
	if cmdData.FinalRepo != nil {
		finalRepo = *cmdData.FinalRepo.Address
	}

	granularSet := len(cacheFrom) > 0 || len(cacheTo) > 0 || len(imagesRepo) > 0 || metaRepo != ""
	aliasSet := len(secondary) > 0 || finalRepo != ""

	if repoSet && (granularSet || aliasSet) {
		return fmt.Errorf("--repo is mutually exclusive with the granular registry flags (--cache-from, --cache-to, --images-repo, --meta-repo) and the deprecated aliases (--secondary-repo, --final-repo); use either --repo as a preset or the granular flags, not both")
	}

	// Deprecation aliases -> new flags.
	if len(secondary) > 0 {
		logboek.Context(ctx).Warn().LogF("DEPRECATED: --secondary-repo is deprecated, use --cache-from instead\n")
		merged := append([]string{}, cacheFrom...)
		cacheFrom = append(merged, secondary...)
		*cmdData.CacheFrom = cacheFrom
		*cmdData.SecondaryStagesStorage = nil
	}
	if finalRepo != "" {
		logboek.Context(ctx).Warn().LogF("DEPRECATED: --final-repo is deprecated, use --images-repo instead\n")
		imagesRepo = append([]string{finalRepo}, imagesRepo...)
		*cmdData.ImagesRepo = imagesRepo
	}

	if opts.ImagesRepoRequired && !repoSet && len(imagesRepo) == 0 {
		return fmt.Errorf("at least one --images-repo=ADDRESS (or --repo=ADDRESS) is required")
	}
	if opts.MetaRepoRequired && !repoSet && metaRepo == "" {
		return fmt.Errorf("--meta-repo=ADDRESS (or --repo=ADDRESS) is required")
	}

	// --cache-from defaults to :local when nothing is configured (and no preset).
	if !repoSet && len(cacheFrom) == 0 {
		*cmdData.CacheFrom = []string{storage.LocalStorageAddress}
	}

	return nil
}
