package git_repo

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dustin/go-humanize"
	"github.com/werf/kubedog/pkg/utils"
	"github.com/werf/lockgate"
	"github.com/werf/logboek"
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
		return false, fmt.Errorf("error getting volume usage by path %q: %s", werf.GetLocalCacheDir(), err)
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

	// TODO: Completely remove v1.1 repos when 1.1 has not been used 2 weeks on the host
	gitReposCacheRoot := filepath.Join(werf.GetLocalCacheDir(), "git_repos")
	if err := wipeCacheDirs(ctx, gitReposCacheRoot, []string{KeepGitRepoCacheVersionV1_1, GitReposCacheVersion}); err != nil {
		return fmt.Errorf("unable to wipe old git repos cache dirs in %q: %s", gitReposCacheRoot, err)
	}

	// TODO: Completely remove v1.1 worktrees when 1.1 has not been used 2 weeks on the host
	gitWorktreesCacheRoot := filepath.Join(werf.GetLocalCacheDir(), "git_worktrees")
	if err := wipeCacheDirs(ctx, gitWorktreesCacheRoot, []string{KeepGitWorkTreeCacheVersionV1_1, GitWorktreesCacheVersion}); err != nil {
		return fmt.Errorf("unable to wipe old git worktrees cache dirs in %q: %s", gitWorktreesCacheRoot, err)
	}

	gitArchivesCacheRoot := filepath.Join(werf.GetLocalCacheDir(), "git_archives")
	if err := wipeCacheDirs(ctx, gitArchivesCacheRoot, []string{GitArchivesCacheVersion}); err != nil {
		return fmt.Errorf("unable to wipe old git archives cache dirs in %q: %s", gitArchivesCacheRoot, err)
	}

	gitPatchesCacheRoot := filepath.Join(werf.GetLocalCacheDir(), "git_patches")
	if err := wipeCacheDirs(ctx, gitPatchesCacheRoot, []string{GitArchivesCacheVersion}); err != nil {
		return fmt.Errorf("unable to wipe old git patches cache dirs in %q: %s", gitPatchesCacheRoot, err)
	}

	// Remove *.tmp from git_repos dir

	vu, err := volumeutils.GetVolumeUsageByPath(ctx, werf.GetLocalCacheDir())
	if err != nil {
		return fmt.Errorf("error getting volume usage by path %q: %s", werf.GetLocalCacheDir(), err)
	}

	targetVolumeUsagePercentage := allowedVolumeUsagePercentage - allowedVolumeUsageMarginPercentage
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

	// var freedBytes uint64
	// TODO: remove following git data based on bytesToFree and LRU
	// TODO: for now temporarily this is complete-wipe-out type of cleanup

	for _, path := range []string{
		filepath.Join(werf.GetLocalCacheDir(), "git_patches", GitPatchesCacheVersion),
		filepath.Join(werf.GetLocalCacheDir(), "git_repos", GitReposCacheVersion),
		filepath.Join(werf.GetLocalCacheDir(), "git_worktrees", GitWorktreesCacheVersion),
		filepath.Join(werf.GetLocalCacheDir(), "git_repos", GitReposCacheVersion),
	} {
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("unable to remove %q: %s", path, err)
		}
	}

	return nil
}

// type GitWorktreeDesc struct {
// }
// TODO: get existing worktrees should gather all existing worktrees and meta info: size and last usage timestamp — this is needed to implement cleanup
// func GetExistingGitWorktrees() ([]*GitWorktreeDesc, error) {
// 	return nil, nil
// }

func wipeCacheDirs(ctx context.Context, cacheRootDir string, keepCacheVersions []string) error {
	if _, err := os.Stat(cacheRootDir); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("error accessing %q: %s", cacheRootDir, err)
	}

	dirs, err := ioutil.ReadDir(cacheRootDir)
	if err != nil {
		return fmt.Errorf("error reading dir %q: %s", cacheRootDir, err)
	}

WipeCacheDirs:
	for _, finfo := range dirs {
		for _, keepCacheVersion := range keepCacheVersions {
			if finfo.Name() == keepCacheVersion {
				continue WipeCacheDirs
			}
		}

		path := filepath.Join(cacheRootDir, finfo.Name())
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("unable to remove %q: %s", path, err)
		}
	}

	return nil
}

func lockGC(ctx context.Context, shared bool) (lockgate.LockHandle, error) {
	_, handle, err := werf.AcquireHostLock(ctx, "git_data_manager", lockgate.AcquireOptions{Shared: shared})
	return handle, err
}
