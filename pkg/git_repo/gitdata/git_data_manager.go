package gitdata

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/werf/lockgate"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

const (
	GitArchivesCacheVersion = "7"
	GitPatchesCacheVersion  = "6"
)

func GetHostGitDataManager(ctx context.Context) (*GitDataManager, error) {
	if lock, err := lockGC(ctx, true); err != nil {
		return nil, err
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	archivesCacheDir := filepath.Join(werf.GetLocalCacheDir(), "git_archives", GitArchivesCacheVersion)
	patchesCacheDir := filepath.Join(werf.GetLocalCacheDir(), "git_patches", GitPatchesCacheVersion)
	tmpGitDataDir := filepath.Join(werf.GetServiceDir(), "tmp", "git_data")

	if err := os.MkdirAll(archivesCacheDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create dir %q: %w", archivesCacheDir, err)
	}
	if err := os.MkdirAll(patchesCacheDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create dir %q: %w", patchesCacheDir, err)
	}
	if err := os.MkdirAll(tmpGitDataDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create dir %q: %w", tmpGitDataDir, err)
	}

	return NewGitDataManager(archivesCacheDir, patchesCacheDir, tmpGitDataDir), nil
}

func NewGitDataManager(archivesCacheDir, patchesCacheDir, tmpDir string) *GitDataManager {
	return &GitDataManager{ArchivesCacheDir: archivesCacheDir, PatchesCacheDir: patchesCacheDir, TmpDir: tmpDir}
}

type GitDataManager struct {
	ArchivesCacheDir string
	PatchesCacheDir  string
	TmpDir           string
}

func (manager *GitDataManager) GetArchivesCacheDir() string {
	return manager.ArchivesCacheDir
}

func (manager *GitDataManager) GetPatchesCacheDir() string {
	return manager.PatchesCacheDir
}

func (manager *GitDataManager) NewTmpFile() (string, error) {
	path := filepath.Join(manager.TmpDir, uuid.NewString())
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		return "", fmt.Errorf("unable to create dir %q: %w", filepath.Dir(path), err)
	}
	return path, nil
}

type PatchMetadata struct {
	Descriptor          *true_git.PatchDescriptor
	LastAccessTimestamp int64
}

type ArchiveMetadata struct {
	LastAccessTimestamp int64
}

func (manager *GitDataManager) GetArchiveFile(ctx context.Context, repoID string, opts git_repo.ArchiveOptions) (*git_repo.ArchiveFile, error) {
	if lock, err := lockGC(ctx, true); err != nil {
		return nil, err
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	if _, lock, err := werf.AcquireHostLock(ctx, fmt.Sprintf("git_archive.%s_%s", repoID, true_git.ArchiveOptions(opts).ID()), lockgate.AcquireOptions{}); err != nil {
		return nil, err
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	metadataPath := filepath.Join(manager.ArchivesCacheDir, archiveMetadataFilePath(repoID, opts))
	if exists, err := util.FileExists(metadataPath); err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}

	path := filepath.Join(manager.ArchivesCacheDir, archiveFilePath(repoID, opts))
	if exists, err := util.FileExists(path); err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}

	if data, err := ioutil.ReadFile(metadataPath); err != nil {
		return nil, fmt.Errorf("error reading %q: %w", metadataPath, err)
	} else {
		var metadata *ArchiveMetadata

		if err := json.Unmarshal(data, &metadata); err != nil {
			return nil, fmt.Errorf("error unmarshalling json from %q: %w", metadataPath, err)
		}

		metadata.LastAccessTimestamp = time.Now().Unix()

		if metadataJson, err := json.Marshal(metadata); err != nil {
			return nil, fmt.Errorf("error marshalling archive %s %s metadata json: %w", repoID, opts.Commit, err)
		} else {
			if err := ioutil.WriteFile(metadataPath, append(metadataJson, '\n'), 0o644); err != nil {
				return nil, fmt.Errorf("error writing %q: %w", metadataPath, err)
			}
		}

		return &git_repo.ArchiveFile{FilePath: path}, nil
	}
}

func (manager *GitDataManager) CreateArchiveFile(ctx context.Context, repoID string, opts git_repo.ArchiveOptions, tmpPath string) (*git_repo.ArchiveFile, error) {
	if archiveFile, err := manager.GetArchiveFile(ctx, repoID, opts); err != nil {
		return nil, err
	} else if archiveFile != nil {
		return archiveFile, nil
	}

	if lock, err := lockGC(ctx, true); err != nil {
		return nil, err
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	if _, lock, err := werf.AcquireHostLock(ctx, fmt.Sprintf("git_archive.%s_%s", repoID, true_git.ArchiveOptions(opts).ID()), lockgate.AcquireOptions{}); err != nil {
		return nil, err
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	metadata := &ArchiveMetadata{
		LastAccessTimestamp: time.Now().Unix(),
	}

	if metadataJson, err := json.Marshal(metadata); err != nil {
		return nil, fmt.Errorf("error marshalling archive %s %s metadata json: %w", repoID, opts.Commit, err)
	} else {
		metadataPath := filepath.Join(manager.ArchivesCacheDir, archiveMetadataFilePath(repoID, opts))
		dir := filepath.Dir(metadataPath)

		if err := os.MkdirAll(dir, 0o777); err != nil {
			return nil, fmt.Errorf("unable to create dir %q: %w", dir, err)
		}

		if err := ioutil.WriteFile(metadataPath, metadataJson, 0o644); err != nil {
			return nil, fmt.Errorf("error writing %s: %w", metadataPath, err)
		}
	}

	path := filepath.Join(manager.ArchivesCacheDir, archiveFilePath(repoID, opts))

	if err := os.Rename(tmpPath, path); err != nil {
		return nil, fmt.Errorf("unable to rename %s to %s: %w", tmpPath, path, err)
	}

	return &git_repo.ArchiveFile{FilePath: path}, nil
}

func (manager *GitDataManager) GetPatchFile(ctx context.Context, repoID string, opts git_repo.PatchOptions) (*git_repo.PatchFile, error) {
	if lock, err := lockGC(ctx, true); err != nil {
		return nil, err
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	if _, lock, err := werf.AcquireHostLock(ctx, fmt.Sprintf("git_patch.%s_%s", repoID, true_git.PatchOptions(opts).ID()), lockgate.AcquireOptions{}); err != nil {
		return nil, err
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	metadataPath := filepath.Join(manager.PatchesCacheDir, patchMetadataFilePath(repoID, opts))
	if exists, err := util.FileExists(metadataPath); err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}

	path := filepath.Join(manager.PatchesCacheDir, patchFilePath(repoID, opts))
	if exists, err := util.FileExists(path); err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}

	if data, err := ioutil.ReadFile(metadataPath); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", metadataPath, err)
	} else {
		var metadata *PatchMetadata

		if err := json.Unmarshal(data, &metadata); err != nil {
			return nil, fmt.Errorf("error unmarshalling json from %s: %w", metadataPath, err)
		}

		metadata.LastAccessTimestamp = time.Now().Unix()

		if metadataJson, err := json.Marshal(metadata); err != nil {
			return nil, fmt.Errorf("error marshalling patch %s %s %s metadata json: %w", repoID, opts.FromCommit, opts.ToCommit, err)
		} else {
			if err := ioutil.WriteFile(metadataPath, append(metadataJson, '\n'), 0o644); err != nil {
				return nil, fmt.Errorf("error writing %s: %w", metadataPath, err)
			}
		}

		return &git_repo.PatchFile{FilePath: path, Descriptor: metadata.Descriptor}, nil
	}
}

func (manager *GitDataManager) LockGC(ctx context.Context, shared bool) (lockgate.LockHandle, error) {
	return lockGC(ctx, shared)
}

func (manager *GitDataManager) CreatePatchFile(ctx context.Context, repoID string, opts git_repo.PatchOptions, tmpPath string, desc *true_git.PatchDescriptor) (*git_repo.PatchFile, error) {
	if patchFile, err := manager.GetPatchFile(ctx, repoID, opts); err != nil {
		return nil, err
	} else if patchFile != nil {
		return patchFile, nil
	}

	if lock, err := lockGC(ctx, true); err != nil {
		return nil, err
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	if _, lock, err := werf.AcquireHostLock(ctx, fmt.Sprintf("git_patch.%s_%s", repoID, true_git.PatchOptions(opts).ID()), lockgate.AcquireOptions{}); err != nil {
		return nil, err
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	metadata := &PatchMetadata{
		Descriptor:          desc,
		LastAccessTimestamp: time.Now().Unix(),
	}

	if metadataJson, err := json.Marshal(metadata); err != nil {
		return nil, fmt.Errorf("error marshalling patch %s %s %s metadata json: %w", repoID, opts.FromCommit, opts.ToCommit, err)
	} else {
		metadataPath := filepath.Join(manager.PatchesCacheDir, patchMetadataFilePath(repoID, opts))
		dir := filepath.Dir(metadataPath)

		if err := os.MkdirAll(dir, 0o777); err != nil {
			return nil, fmt.Errorf("unable to create dir %q: %w", dir, err)
		}

		if err := ioutil.WriteFile(metadataPath, append(metadataJson, '\n'), 0o644); err != nil {
			return nil, fmt.Errorf("error writing %s: %w", metadataPath, err)
		}
	}

	path := filepath.Join(manager.PatchesCacheDir, patchFilePath(repoID, opts))

	if err := os.Rename(tmpPath, path); err != nil {
		return nil, fmt.Errorf("unable to rename %s to %s: %w", tmpPath, path, err)
	}

	return &git_repo.PatchFile{FilePath: path, Descriptor: desc}, nil
}

func patchMetadataFilePath(repoID string, opts git_repo.PatchOptions) string {
	return fmt.Sprintf("%s.meta.json", commonGitDataFilePath(repoID, true_git.PatchOptions(opts).ID()))
}

func patchFilePath(repoID string, opts git_repo.PatchOptions) string {
	return fmt.Sprintf("%s.patch", commonGitDataFilePath(repoID, true_git.PatchOptions(opts).ID()))
}

func archiveMetadataFilePath(repoID string, opts git_repo.ArchiveOptions) string {
	return fmt.Sprintf("%s.meta.json", commonGitDataFilePath(repoID, true_git.ArchiveOptions(opts).ID()))
}

func archiveFilePath(repoID string, opts git_repo.ArchiveOptions) string {
	return fmt.Sprintf("%s.tar", commonGitDataFilePath(repoID, true_git.ArchiveOptions(opts).ID()))
}

func commonGitDataFilePath(repoID, id string) string {
	return fmt.Sprintf("%s/%s/%s", repoID, id[0:2], id)
}
