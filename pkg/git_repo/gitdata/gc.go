package gitdata

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/kubedog/pkg/utils"
	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	thresholdpkg "github.com/werf/werf/v2/pkg/cleaning/threshold"
	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/volumeutils"
	"github.com/werf/werf/v2/pkg/werf"
)

const (
	KeepGitWorkTreeCacheVersionV1_1 = "6"
	KeepGitRepoCacheVersionV1_1     = "3"
)

func exceedsLocalCacheVolumeUsageThreshold(vu volumeutils.VolumeUsage, threshold thresholdpkg.Threshold) bool {
	switch threshold.Type {
	case thresholdpkg.TypePercentage:
		return vu.Percentage() > threshold.PercentageValue()
	case thresholdpkg.TypeBytes:
		return vu.FreeBytes() < threshold.Value
	default:
		panic(fmt.Sprintf("unexpected volume usage threshold type %q", threshold.Type))
	}
}

func targetLocalCacheBytesToFree(vu volumeutils.VolumeUsage, threshold, margin thresholdpkg.Threshold) uint64 {
	switch threshold.Type {
	case thresholdpkg.TypePercentage:
		return vu.BytesToFree(max(threshold.PercentageValue()-margin.PercentageValue(), 0))
	case thresholdpkg.TypeBytes:
		return vu.BytesToFreeForTargetFreeBytes(threshold.Value + margin.Value)
	default:
		panic(fmt.Sprintf("unexpected volume usage threshold type %q", threshold.Type))
	}
}

func logLocalCacheVolumeUsageThresholdCheck(ctx context.Context, vu volumeutils.VolumeUsage, threshold thresholdpkg.Threshold) {
	switch threshold.Type {
	case thresholdpkg.TypePercentage:
		logboek.Context(ctx).Default().LogF("Allowed volume usage percentage: %s <= %s — %s\n", utils.GreenF("%0.2f%%", vu.Percentage()), utils.BlueF("%0.2f%%", threshold.PercentageValue()), utils.GreenF("OK"))
	case thresholdpkg.TypeBytes:
		logboek.Context(ctx).Default().LogF("Allowed free space: %s >= %s — %s\n", utils.GreenF("%s", humanize.Bytes(vu.FreeBytes())), utils.BlueF("%s", humanize.Bytes(threshold.Value)), utils.GreenF("OK"))
	default:
		panic(fmt.Sprintf("unexpected volume usage threshold type %q", threshold.Type))
	}
}

func logExceededLocalCacheVolumeUsageThreshold(ctx context.Context, vu volumeutils.VolumeUsage, threshold thresholdpkg.Threshold) {
	switch threshold.Type {
	case thresholdpkg.TypePercentage:
		logboek.Context(ctx).Default().LogF("Allowed percentage level exceeded: %s > %s — %s\n", utils.RedF("%0.2f%%", vu.Percentage()), utils.YellowF("%0.2f%%", threshold.PercentageValue()), utils.RedF("HIGH VOLUME USAGE"))
	case thresholdpkg.TypeBytes:
		logboek.Context(ctx).Default().LogF("Allowed free space level exceeded: %s < %s — %s\n", utils.RedF("%s", humanize.Bytes(vu.FreeBytes())), utils.YellowF("%s", humanize.Bytes(threshold.Value)), utils.RedF("LOW FREE SPACE"))
	default:
		panic(fmt.Sprintf("unexpected volume usage threshold type %q", threshold.Type))
	}
}

func logTargetLocalCacheVolumeUsageThreshold(ctx context.Context, threshold, margin thresholdpkg.Threshold, bytesToFree uint64) {
	switch threshold.Type {
	case thresholdpkg.TypePercentage:
		logboek.Context(ctx).Default().LogF("Target percentage level after cleanup: %0.2f%% - %0.2f%% (margin) = %s\n", threshold.PercentageValue(), margin.PercentageValue(), utils.BlueF("%0.2f%%", max(threshold.PercentageValue()-margin.PercentageValue(), 0)))
	case thresholdpkg.TypeBytes:
		logboek.Context(ctx).Default().LogF("Target free space after cleanup: %s + %s (margin) = %s\n", humanize.Bytes(threshold.Value), humanize.Bytes(margin.Value), utils.BlueF("%s", humanize.Bytes(threshold.Value+margin.Value)))
	default:
		panic(fmt.Sprintf("unexpected volume usage threshold type %q", threshold.Type))
	}
	logboek.Context(ctx).Default().LogF("Needed to free: %s\n", utils.RedF("%s", humanize.Bytes(bytesToFree)))
}

func ShouldRunAutoGC(ctx context.Context, allowedVolumeUsageThreshold thresholdpkg.Threshold) (bool, error) {
	vu, err := volumeutils.GetVolumeUsageByPath(ctx, werf.GetLocalCacheDir())
	if err != nil {
		return false, fmt.Errorf("error getting volume usage by path %q: %w", werf.GetLocalCacheDir(), err)
	}

	return exceedsLocalCacheVolumeUsageThreshold(vu, allowedVolumeUsageThreshold), nil
}

type RunGCOptions struct {
	AllowedLocalCacheVolumeUsageThreshold       thresholdpkg.Threshold
	AllowedLocalCacheVolumeUsageMarginThreshold thresholdpkg.Threshold
	DryRun                                      bool
}

func RunGC(ctx context.Context, options RunGCOptions) error {
	if lock, err := lockGC(ctx, false); err != nil {
		return err
	} else {
		defer werf.HostLocker().ReleaseLock(lock)
	}

	keepGitDataV1_1, err := shouldKeepGitDataV1_1(ctx)
	if err != nil {
		return fmt.Errorf("unable to check if git data v1.1 should be kept: %w", err)
	}

	{
		keepCacheVersions := []string{git_repo.GitReposCacheVersion}
		if keepGitDataV1_1 {
			keepCacheVersions = append(keepCacheVersions, KeepGitRepoCacheVersionV1_1)
		}

		cacheRoot := filepath.Join(werf.GetLocalCacheDir(), "git_repos")
		if err := wipeCacheDirs(ctx, cacheRoot, keepCacheVersions); err != nil {
			return fmt.Errorf("unable to wipe old git repos cache dirs in %q: %w", cacheRoot, err)
		}
	}

	{
		keepCacheVersions := []string{git_repo.GitWorktreesCacheVersion}
		if keepGitDataV1_1 {
			keepCacheVersions = append(keepCacheVersions, KeepGitWorkTreeCacheVersionV1_1)
		}

		cacheRoot := filepath.Join(werf.GetLocalCacheDir(), "git_worktrees")
		if err := wipeCacheDirs(ctx, cacheRoot, keepCacheVersions); err != nil {
			return fmt.Errorf("unable to wipe old git worktrees cache dirs in %q: %w", cacheRoot, err)
		}
	}

	{
		cacheRoot := filepath.Join(werf.GetLocalCacheDir(), "git_archives")
		if err := wipeCacheDirs(ctx, cacheRoot, []string{GitArchivesCacheVersion}); err != nil {
			return fmt.Errorf("unable to wipe old git archives cache dirs in %q: %w", cacheRoot, err)
		}
	}

	{
		cacheRoot := filepath.Join(werf.GetLocalCacheDir(), "git_patches")
		if err := wipeCacheDirs(ctx, cacheRoot, []string{GitPatchesCacheVersion}); err != nil {
			return fmt.Errorf("unable to wipe old git patches cache dirs in %q: %w", cacheRoot, err)
		}
	}

	vu, err := volumeutils.GetVolumeUsageByPath(ctx, werf.GetLocalCacheDir())
	if err != nil {
		return fmt.Errorf("error getting volume usage by path %q: %w", werf.GetLocalCacheDir(), err)
	}

	bytesToFree := targetLocalCacheBytesToFree(vu, options.AllowedLocalCacheVolumeUsageThreshold, options.AllowedLocalCacheVolumeUsageMarginThreshold)

	if !exceedsLocalCacheVolumeUsageThreshold(vu, options.AllowedLocalCacheVolumeUsageThreshold) {
		logboek.Context(ctx).Default().LogBlock("Git data storage check").Do(func() {
			logboek.Context(ctx).Default().LogF("Werf local cache dir: %s\n", werf.GetLocalCacheDir())
			logboek.Context(ctx).Default().LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
			logLocalCacheVolumeUsageThresholdCheck(ctx, vu, options.AllowedLocalCacheVolumeUsageThreshold)
		})

		return nil
	}

	logboek.Context(ctx).Default().LogBlock("Git data storage check").Do(func() {
		logboek.Context(ctx).Default().LogF("Werf local cache dir: %s\n", werf.GetLocalCacheDir())
		logboek.Context(ctx).Default().LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
		logExceededLocalCacheVolumeUsageThreshold(ctx, vu, options.AllowedLocalCacheVolumeUsageThreshold)
		logTargetLocalCacheVolumeUsageThreshold(ctx, options.AllowedLocalCacheVolumeUsageThreshold, options.AllowedLocalCacheVolumeUsageMarginThreshold, bytesToFree)
	})

	var gitDataEntries []GitDataEntry

	{
		cacheVersionRoot := filepath.Join(werf.GetLocalCacheDir(), "git_repos", git_repo.GitReposCacheVersion)

		entries, err := GetGitReposAndRemoveInvalid(ctx, cacheVersionRoot)
		if err != nil {
			return fmt.Errorf("unable to process git repos from %q: %w", cacheVersionRoot, err)
		}

		gitDataEntries = append(gitDataEntries, entries...)
	}

	{
		cacheVersionRoot := filepath.Join(werf.GetLocalCacheDir(), "git_worktrees", git_repo.GitWorktreesCacheVersion)

		entries, err := GetGitWorktreesAndRemoveInvalid(ctx, cacheVersionRoot)
		if err != nil {
			return fmt.Errorf("unable to process git worktrees from %q: %w", cacheVersionRoot, err)
		}

		gitDataEntries = append(gitDataEntries, entries...)
	}

	{
		cacheVersionRoot := filepath.Join(werf.GetLocalCacheDir(), "git_archives", GitArchivesCacheVersion)

		entries, err := GetGitArchivesAndRemoveInvalid(ctx, cacheVersionRoot)
		if err != nil {
			return fmt.Errorf("unable to process git archives from %q: %w", cacheVersionRoot, err)
		}

		gitDataEntries = append(gitDataEntries, entries...)
	}

	{
		cacheVersionRoot := filepath.Join(werf.GetLocalCacheDir(), "git_patches", GitPatchesCacheVersion)

		entries, err := GetGitPatchesAndRemoveInvalid(ctx, cacheVersionRoot)
		if err != nil {
			return fmt.Errorf("unable to process git patches from %q: %w", cacheVersionRoot, err)
		}

		gitDataEntries = append(gitDataEntries, entries...)
	}

	gitDataEntries = keepGitDataByLru(gitDataEntries)

	var freedBytes uint64

	for _, entry := range gitDataEntries {
		for _, path := range entry.GetPaths() {
			logboek.Context(ctx).LogF("Removing %q inside scope %q\n", path, entry.GetCacheBasePath())

			if options.DryRun {
				continue
			}

			if err := RemovePathWithEmptyParentDirsInsideScope(entry.GetCacheBasePath(), path); err != nil {
				return fmt.Errorf("unable to remove %q: %w", path, err)
			}
		}

		freedBytes += entry.GetSize()

		if freedBytes >= bytesToFree {
			break
		}
	}

	return nil
}

func RemovePathWithEmptyParentDirsInsideScope(scopeDir, path string) error {
	if !util.IsSubpathOfBasePath(scopeDir, path) {
		return nil
	}

	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("unable to remove %q: %w", path, err)
	}

	dir := filepath.Dir(path)

	for {
		if !util.IsSubpathOfBasePath(scopeDir, dir) {
			break
		}

		files, err := ioutil.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("error reading dir %q: %s", dir, files)
		}

		if len(files) > 0 {
			break
		}

		if err := os.Remove(dir); err != nil {
			return fmt.Errorf("unable to remove empty dir %q: %w", dir, err)
		}

		dir = filepath.Dir(dir)
	}

	return nil
}

// wipeCacheDirs removes all subdirectories from the specified cache root directory,
// except for those listed in keepCacheVersions.
func wipeCacheDirs(ctx context.Context, cacheRootDir string, keepCacheVersions []string) error {
	logboek.Context(ctx).Debug().LogF("wipeCacheDirs %q\n", cacheRootDir)

	if _, err := os.Stat(cacheRootDir); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("error accessing %q: %w", cacheRootDir, err)
	}

	dirs, err := ioutil.ReadDir(cacheRootDir)
	if err != nil {
		return fmt.Errorf("error reading dir %q: %w", cacheRootDir, err)
	}

	for _, finfo := range dirs {
		if slices.Contains(keepCacheVersions, finfo.Name()) {
			logboek.Context(ctx).Debug().LogF("wipeCacheDirs in %q: keep cache version %q\n", cacheRootDir, finfo.Name())
			continue
		}

		versionedCacheDir := filepath.Join(cacheRootDir, finfo.Name())

		if err = os.RemoveAll(versionedCacheDir); err != nil {
			return fmt.Errorf("unable to remove %q: %w", versionedCacheDir, err)
		}
	}

	return nil
}

func lockGC(ctx context.Context, shared bool) (lockgate.LockHandle, error) {
	_, handle, err := werf.HostLocker().AcquireLock(ctx, "git_data_manager", lockgate.AcquireOptions{Shared: shared})
	return handle, err
}

// shouldKeepGitDataV1_1 returns true if the last run of werf v1.1 was within the last 3 days.
func shouldKeepGitDataV1_1(ctx context.Context) (bool, error) {
	v1_1LastRunAt, err := werf.GetWerfLastRunAtV1_1(ctx)
	if err != nil {
		return false, fmt.Errorf("error getting last run timestamp for werf v1.1: %w", err)
	}

	return time.Since(v1_1LastRunAt) <= time.Hour*24*3, nil
}
