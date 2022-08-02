package gitdata

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/werf/kubedog/pkg/utils"
	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/volumeutils"
	"github.com/werf/werf/pkg/werf"
)

const (
	KeepGitWorkTreeCacheVersionV1_1 = "6"
	KeepGitRepoCacheVersionV1_1     = "3"
)

func ShouldRunAutoGC(ctx context.Context, allowedVolumeUsagePercentage float64) (bool, error) {
	vu, err := volumeutils.GetVolumeUsageByPath(ctx, werf.GetLocalCacheDir())
	if err != nil {
		return false, fmt.Errorf("error getting volume usage by path %q: %w", werf.GetLocalCacheDir(), err)
	}

	return vu.Percentage > allowedVolumeUsagePercentage, nil
}

func getBytesToFree(vu volumeutils.VolumeUsage, targetVolumeUsagePercentage float64) uint64 {
	allowedVolumeUsageToFree := vu.Percentage - targetVolumeUsagePercentage
	return uint64((float64(vu.TotalBytes) / 100.0) * allowedVolumeUsageToFree)
}

func RunGC(ctx context.Context, allowedVolumeUsagePercentage, allowedVolumeUsageMarginPercentage float64) error {
	if lock, err := lockGC(ctx, false); err != nil {
		return err
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	var keepGitDataV1_1 bool
	v1_1LastRunAt, err := werf.GetWerfLastRunAtV1_1(ctx)
	if err != nil {
		return fmt.Errorf("error getting last run timestamp for werf v1.1: %w", err)
	}
	if time.Since(v1_1LastRunAt) <= time.Hour*24*3 {
		keepGitDataV1_1 = true
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

		for _, dir := range []string{filepath.Join(cacheRoot, git_repo.GitReposCacheVersion), filepath.Join(cacheRoot, KeepGitRepoCacheVersionV1_1)} {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				continue
			} else if err != nil {
				return fmt.Errorf("error accessing dir %q: %w", dir, err)
			}

			files, err := ioutil.ReadDir(dir)
			if err != nil {
				return fmt.Errorf("error reading dir %q: %w", dir, err)
			}

			for _, finfo := range files {
				if strings.HasSuffix(finfo.Name(), ".tmp") {
					path := filepath.Join(dir, finfo.Name())
					if err := os.RemoveAll(path); err != nil {
						return fmt.Errorf("unable to remove %q: %w", path, err)
					}
				}
			}
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

	targetVolumeUsagePercentage := allowedVolumeUsagePercentage - allowedVolumeUsageMarginPercentage
	if targetVolumeUsagePercentage < 0 {
		targetVolumeUsagePercentage = 0
	}

	bytesToFree := getBytesToFree(vu, targetVolumeUsagePercentage)

	if vu.Percentage <= allowedVolumeUsagePercentage {
		logboek.Context(ctx).Default().LogBlock("Git data storage check").Do(func() {
			logboek.Context(ctx).Default().LogF("Werf local cache dir: %s\n", werf.GetLocalCacheDir())
			logboek.Context(ctx).Default().LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
			logboek.Context(ctx).Default().LogF("Allowed volume usage percentage: %s <= %s — %s\n", utils.GreenF("%0.2f%%", vu.Percentage), utils.BlueF("%0.2f%%", allowedVolumeUsagePercentage), utils.GreenF("OK"))
		})

		return nil
	}

	logboek.Context(ctx).Default().LogBlock("Git data storage check").Do(func() {
		logboek.Context(ctx).Default().LogF("Werf local cache dir: %s\n", werf.GetLocalCacheDir())
		logboek.Context(ctx).Default().LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
		logboek.Context(ctx).Default().LogF("Allowed percentage level exceeded: %s > %s — %s\n", utils.RedF("%0.2f%%", vu.Percentage), utils.YellowF("%0.2f%%", allowedVolumeUsagePercentage), utils.RedF("HIGH VOLUME USAGE"))
		logboek.Context(ctx).Default().LogF("Target percentage level after cleanup: %0.2f%% - %0.2f%% (margin) = %s\n", allowedVolumeUsagePercentage, allowedVolumeUsageMarginPercentage, utils.BlueF("%0.2f%%", targetVolumeUsagePercentage))
		logboek.Context(ctx).Default().LogF("Needed to free: %s\n", utils.RedF("%s", humanize.Bytes(bytesToFree)))
	})

	var gitDataEntries []GitDataEntry

	{
		cacheVersionRoot := filepath.Join(werf.GetLocalCacheDir(), "git_repos", git_repo.GitReposCacheVersion)

		entries, err := GetExistingGitRepos(cacheVersionRoot)
		if err != nil {
			return fmt.Errorf("error getting existing git repos from %q: %w", cacheVersionRoot, err)
		}

		for _, entry := range entries {
			gitDataEntries = append(gitDataEntries, entry)
		}
	}

	{
		cacheVersionRoot := filepath.Join(werf.GetLocalCacheDir(), "git_worktrees", git_repo.GitWorktreesCacheVersion)

		entries, err := GetExistingGitWorktrees(cacheVersionRoot)
		if err != nil {
			return fmt.Errorf("error getting existing git repos from %q: %w", cacheVersionRoot, err)
		}

		for _, entry := range entries {
			gitDataEntries = append(gitDataEntries, entry)
		}
	}

	{
		cacheVersionRoot := filepath.Join(werf.GetLocalCacheDir(), "git_archives", GitArchivesCacheVersion)

		entries, err := GetExistingGitArchives(cacheVersionRoot)
		if err != nil {
			return fmt.Errorf("error getting existing git repos from %q: %w", cacheVersionRoot, err)
		}

		for _, entry := range entries {
			gitDataEntries = append(gitDataEntries, entry)
		}
	}

	{
		cacheVersionRoot := filepath.Join(werf.GetLocalCacheDir(), "git_patches", GitPatchesCacheVersion)

		entries, err := GetExistingGitPatches(cacheVersionRoot)
		if err != nil {
			return fmt.Errorf("error getting existing git repos from %q: %w", cacheVersionRoot, err)
		}

		for _, entry := range entries {
			gitDataEntries = append(gitDataEntries, entry)
		}
	}

	sort.Sort(GitDataLruSort(gitDataEntries))

	gitDataEntries = PreserveGitDataByLru(gitDataEntries)

	var freedBytes uint64
	for _, entry := range gitDataEntries {
		for _, path := range entry.GetPaths() {
			logboek.Context(ctx).LogF("Removing %q inside scope %q\n", path, entry.GetCacheBasePath())
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

WipeCacheDirs:
	for _, finfo := range dirs {
		for _, keepCacheVersion := range keepCacheVersions {
			if finfo.Name() == keepCacheVersion {
				logboek.Context(ctx).Debug().LogF("wipeCacheDirs in %q: keep cache version %q\n", cacheRootDir, keepCacheVersion)
				continue WipeCacheDirs
			}
		}

		versionedCacheDir := filepath.Join(cacheRootDir, finfo.Name())
		subdirs, err := ioutil.ReadDir(versionedCacheDir)
		if err != nil {
			return fmt.Errorf("error reading dir %q: %w", versionedCacheDir, err)
		}

		for _, cacheFile := range subdirs {
			path := filepath.Join(versionedCacheDir, cacheFile.Name())

			logboek.Context(ctx).Debug().LogF("wipeCacheDirs in %q: removing %q\n", cacheRootDir, path)
			if err := os.RemoveAll(path); err != nil {
				return fmt.Errorf("unable to remove %q: %w", path, err)
			}
		}
	}

	return nil
}

func lockGC(ctx context.Context, shared bool) (lockgate.LockHandle, error) {
	_, handle, err := werf.AcquireHostLock(ctx, "git_data_manager", lockgate.AcquireOptions{Shared: shared})
	return handle, err
}
