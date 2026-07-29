package gitdata

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/samber/lo"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/common-go/pkg/util/timestamps"
	"github.com/werf/kubedog/pkg/utils"
	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/volumeutils"
	"github.com/werf/werf/v2/pkg/werf"
)

const (
	KeepGitWorkTreeCacheVersionV1_1 = "6"
	KeepGitRepoCacheVersionV1_1     = "3"
)

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

	keepGitDataV1_1, err := shouldKeepGitDataV1_1(ctx)
	if err != nil {
		return fmt.Errorf("unable to check if git data v1.1 should be kept: %w", err)
	}

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

	for _, foreignVersionsDesc := range []struct {
		cacheRoot    string
		keepVersions []string
	}{
		{filepath.Join(werf.GetLocalCacheDir(), "git_repos"), lo.Ternary(keepGitDataV1_1, []string{git_repo.GitReposCacheVersion, KeepGitRepoCacheVersionV1_1}, []string{git_repo.GitReposCacheVersion})},
		{filepath.Join(werf.GetLocalCacheDir(), "git_worktrees"), lo.Ternary(keepGitDataV1_1, []string{git_repo.GitWorktreesCacheVersion, KeepGitWorkTreeCacheVersionV1_1}, []string{git_repo.GitWorktreesCacheVersion})},
		{filepath.Join(werf.GetLocalCacheDir(), "git_archives"), []string{GitArchivesCacheVersion}},
		{filepath.Join(werf.GetLocalCacheDir(), "git_patches"), []string{GitPatchesCacheVersion}},
	} {
		entries, err := getForeignCacheVersionEntries(foreignVersionsDesc.cacheRoot, foreignVersionsDesc.keepVersions)
		if err != nil {
			return fmt.Errorf("unable to process foreign cache versions in %q: %w", foreignVersionsDesc.cacheRoot, err)
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

	// Rename before removal so that concurrent readers see the path disappear atomically
	// instead of observing a partially deleted directory.
	removePath := fmt.Sprintf("%s.removing.%d", path, time.Now().UnixNano())
	if err := os.Rename(path, removePath); err != nil {
		removePath = path
	}

	if err := os.RemoveAll(removePath); err != nil {
		return fmt.Errorf("unable to remove %q: %w", removePath, err)
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

type foreignCacheVersionDesc struct {
	Path          string
	LastAccessAt  time.Time
	Size          uint64
	CacheBasePath string
}

var _ GitDataEntry = (*foreignCacheVersionDesc)(nil)

func (entry *foreignCacheVersionDesc) GetPaths() []string {
	return []string{entry.Path}
}

func (entry *foreignCacheVersionDesc) GetSize() uint64 {
	return entry.Size
}

func (entry *foreignCacheVersionDesc) GetLastAccessAt() time.Time {
	return entry.LastAccessAt
}

func (entry *foreignCacheVersionDesc) GetCacheBasePath() string {
	return entry.CacheBasePath
}

// getForeignCacheVersionEntries returns LRU entries for cache version dirs maintained by other
// werf versions running on the same host. Each foreign version dir is a single entry: its layout
// is unknown to the current werf version, so it is only measured, never validated or modified.
// Last access time is the newest last_access_at timestamp found near the top of the dir, falling
// back to the dir modification time.
func getForeignCacheVersionEntries(cacheRootDir string, keepVersions []string) ([]GitDataEntry, error) {
	dirs, err := ioutil.ReadDir(cacheRootDir)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("error reading dir %q: %w", cacheRootDir, err)
	}

	var res []GitDataEntry

	for _, versionDirInfo := range dirs {
		if !versionDirInfo.IsDir() || slices.Contains(keepVersions, versionDirInfo.Name()) {
			continue
		}

		versionDir := filepath.Join(cacheRootDir, versionDirInfo.Name())

		size, err := volumeutils.DirSizeBytes(versionDir)
		if err != nil {
			return nil, fmt.Errorf("error getting dir %q size: %w", versionDir, err)
		}

		lastAccessAt, found := findMaxLastAccessAt(versionDir, 2)
		if !found {
			lastAccessAt = versionDirInfo.ModTime()
		}

		res = append(res, &foreignCacheVersionDesc{
			Path:          versionDir,
			LastAccessAt:  lastAccessAt,
			Size:          size,
			CacheBasePath: cacheRootDir,
		})
	}

	return res, nil
}

func findMaxLastAccessAt(dir string, depth int) (time.Time, bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return time.Time{}, false
	}

	var res time.Time
	var found bool

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			if depth > 0 {
				if t, ok := findMaxLastAccessAt(path, depth-1); ok && t.After(res) {
					res = t
					found = true
				}
			}
			continue
		}

		if entry.Name() != "last_access_at" {
			continue
		}

		if t, err := timestamps.ReadTimestampFile(path); err == nil && t.After(res) {
			res = t
			found = true
		}
	}

	return res, found
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
