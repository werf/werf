package gitdata

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/samber/lo"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/kubedog/pkg/utils"
	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/volumeutils"
	"github.com/werf/werf/v2/pkg/werf"
)

const cacheVersionStalenessWindow = 3 * 24 * time.Hour

var errFreshFileFound = errors.New("fresh file found")

func ShouldRunAutoGC(ctx context.Context, allowedVolumeUsageBytes uint64) (bool, error) {
	vu, err := volumeutils.GetVolumeUsageByPath(ctx, werf.GetLocalCacheDir())
	if err != nil {
		return false, fmt.Errorf("error getting volume usage by path %q: %w", werf.GetLocalCacheDir(), err)
	}

	return vu.UsedBytes > allowedVolumeUsageBytes, nil
}

type RunGCOptions struct {
	AllowedLocalCacheVolumeUsageBytes       uint64
	AllowedLocalCacheVolumeUsageMarginBytes uint64
	DryRun                                  bool
}

func RunGC(ctx context.Context, options RunGCOptions) error {
	if lock, err := lockGC(ctx, false); err != nil {
		return err
	} else {
		defer werf.HostLocker().ReleaseLock(lock)
	}

	{
		cacheRoot := filepath.Join(werf.GetLocalCacheDir(), "git_repos")
		if err := wipeCacheDirs(ctx, cacheRoot, []string{git_repo.GitReposCacheVersion}); err != nil {
			return fmt.Errorf("unable to wipe old git repos cache dirs in %q: %w", cacheRoot, err)
		}
	}

	{
		cacheRoot := filepath.Join(werf.GetLocalCacheDir(), "git_mirrors")
		if err := wipeCacheDirs(ctx, cacheRoot, []string{git_repo.GitMirrorsCacheVersion}); err != nil {
			return fmt.Errorf("unable to wipe old git mirrors cache dirs in %q: %w", cacheRoot, err)
		}
	}

	{
		cacheRoot := filepath.Join(werf.GetLocalCacheDir(), "git_worktrees")
		if err := wipeCacheDirs(ctx, cacheRoot, []string{git_repo.GitWorktreesCacheVersion}); err != nil {
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

	if vu.UsedBytes <= options.AllowedLocalCacheVolumeUsageBytes {
		logboek.Context(ctx).Default().LogBlock("Git data storage check").Do(func() {
			logboek.Context(ctx).Default().LogF("Werf local cache dir: %s\n", werf.GetLocalCacheDir())
			logboek.Context(ctx).Default().LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
			logboek.Context(ctx).Default().LogF("Allowed volume usage: %s <= %s — %s\n", utils.GreenF("%s (%.2f%%)", humanize.Bytes(vu.UsedBytes), vu.BytesToPercentage(vu.UsedBytes)), utils.BlueF("%s (%.2f%%)", humanize.Bytes(options.AllowedLocalCacheVolumeUsageBytes), vu.BytesToPercentage(options.AllowedLocalCacheVolumeUsageBytes)), utils.GreenF("OK"))
		})

		return nil
	}

	targetVolumeUsageBytes := uint64(math.Max(float64(options.AllowedLocalCacheVolumeUsageBytes)-float64(options.AllowedLocalCacheVolumeUsageMarginBytes), 0))
	bytesToFree := lo.Ternary(vu.UsedBytes > targetVolumeUsageBytes, vu.UsedBytes-targetVolumeUsageBytes, 0)

	logboek.Context(ctx).Default().LogBlock("Git data storage check").Do(func() {
		logboek.Context(ctx).Default().LogF("Werf local cache dir: %s\n", werf.GetLocalCacheDir())
		logboek.Context(ctx).Default().LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
		logboek.Context(ctx).Default().LogF("Allowed level exceeded: %s > %s — %s\n", utils.RedF("%s (%.2f%%)", humanize.Bytes(vu.UsedBytes), vu.BytesToPercentage(vu.UsedBytes)), utils.YellowF("%s (%.2f%%)", humanize.Bytes(options.AllowedLocalCacheVolumeUsageBytes), vu.BytesToPercentage(options.AllowedLocalCacheVolumeUsageBytes)), utils.RedF("HIGH VOLUME USAGE"))
		logboek.Context(ctx).Default().LogF("Target level after cleanup: %s - %s (margin) = %s\n", humanize.Bytes(options.AllowedLocalCacheVolumeUsageBytes), humanize.Bytes(options.AllowedLocalCacheVolumeUsageMarginBytes), utils.BlueF("%s (%.2f%%)", humanize.Bytes(targetVolumeUsageBytes), vu.BytesToPercentage(targetVolumeUsageBytes)))
		logboek.Context(ctx).Default().LogF("Needed to free: %s\n", utils.RedF("%s", humanize.Bytes(bytesToFree)))
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
		cacheVersionRoot := filepath.Join(werf.GetLocalCacheDir(), "git_mirrors", git_repo.GitMirrorsCacheVersion)

		entries, err := GetGitMirrorsAndRemoveInvalid(ctx, cacheVersionRoot)
		if err != nil {
			return fmt.Errorf("unable to process git mirrors from %q: %w", cacheVersionRoot, err)
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

// wipeCacheDirs removes non-current version dirs from cacheRootDir, but only
// stale ones: local_cache is shared by all werf versions on the host (see
// package doc), and a version dir with recent file mtimes belongs to another
// werf version that may be mid-build right now.
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

		fresh, err := hasFreshFiles(versionedCacheDir)
		if err != nil {
			logboek.Context(ctx).Warn().LogF("Keeping %q: unable to check for fresh files: %s\n", versionedCacheDir, err)
			continue
		}
		if fresh {
			logboek.Context(ctx).Debug().LogF("wipeCacheDirs in %q: keep fresh %q\n", cacheRootDir, finfo.Name())
			continue
		}

		if err = os.RemoveAll(versionedCacheDir); err != nil {
			return fmt.Errorf("unable to remove %q: %w", versionedCacheDir, err)
		}
	}

	return nil
}

func hasFreshFiles(path string) (bool, error) {
	err := filepath.WalkDir(path, func(_ string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		if !dirEntry.Type().IsRegular() {
			return nil
		}

		info, err := dirEntry.Info()
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		if time.Since(info.ModTime()) <= cacheVersionStalenessWindow {
			return errFreshFileFound
		}

		return nil
	})
	if errors.Is(err, errFreshFileFound) {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	return false, nil
}

func lockGC(ctx context.Context, shared bool) (lockgate.LockHandle, error) {
	_, handle, err := werf.HostLocker().AcquireLock(ctx, "git_data_manager", lockgate.AcquireOptions{Shared: shared})
	return handle, err
}
