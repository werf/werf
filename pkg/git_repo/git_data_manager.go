package git_repo

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/werf/werf/pkg/true_git"

	"github.com/werf/werf/pkg/util"

	"github.com/werf/lockgate"

	uuid "github.com/satori/go.uuid"
	"github.com/werf/werf/pkg/werf"
)

const (
	GitArchivesCacheVersion = "3"
	GitPatchesCacheVersion  = "3"
)

var (
	CommonGitDataManager *GitDataManager
)

func Init() error {
	CommonGitDataManager = NewGitDataManager(
		filepath.Join(werf.GetLocalCacheDir(), "git_archives", GitArchivesCacheVersion),
		filepath.Join(werf.GetLocalCacheDir(), "git_patches", GitPatchesCacheVersion),
		filepath.Join(werf.GetServiceDir(), "tmp", "git_data"),
	)
	return nil
}

func NewGitDataManager(archivesCacheDir, patchesCacheDir, tmpDir string) *GitDataManager {
	return &GitDataManager{ArchivesCacheDir: archivesCacheDir, PatchesCacheDir: patchesCacheDir, TmpDir: tmpDir}
}

type GitDataManager struct {
	ArchivesCacheDir string
	PatchesCacheDir  string
	TmpDir           string
}

func (manager *GitDataManager) GC() error {
	return nil
}

func (manager *GitDataManager) getArchiveCacheFilePath(repoID, commit string) string {
	return filepath.Join(manager.ArchivesCacheDir, repoID, fmt.Sprintf("%s.tar", commit))
}

func (manager *GitDataManager) NewTmpFile() (string, error) {
	path := filepath.Join(manager.TmpDir, uuid.NewV4().String())
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return "", fmt.Errorf("unable to create dir %q: %s", filepath.Dir(path), err)
	}
	return path, nil
}

func (manager *GitDataManager) GetArchiveFile(ctx context.Context, repoID string, opts ArchiveOptions) (*ArchiveFile, error) {
	if lock, err := manager.lockGC(ctx, true); err != nil {
		return nil, err
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	metadataPath := filepath.Join(manager.ArchivesCacheDir, archiveMetadataFileName(repoID, opts))
	if exists, err := util.FileExists(metadataPath); err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}

	path := filepath.Join(manager.ArchivesCacheDir, archiveFileName(repoID, opts))
	if exists, err := util.FileExists(path); err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}

	if data, err := ioutil.ReadFile(metadataPath); err != nil {
		return nil, fmt.Errorf("error reading %s: %s", metadataPath, err)
	} else {
		var desc *true_git.ArchiveDescriptor
		if err := json.Unmarshal(data, &desc); err != nil {
			return nil, fmt.Errorf("error unmarshalling json from %s: %s", metadataPath, err)
		}

		return &ArchiveFile{FilePath: path, Descriptor: desc}, nil
	}
}

func (manager *GitDataManager) lockGC(ctx context.Context, readOnly bool) (lockgate.LockHandle, error) {
	_, handle, err := werf.AcquireHostLock(ctx, "git_data_manager", lockgate.AcquireOptions{NonBlocking: readOnly})
	return handle, err
}

func (manager *GitDataManager) CreateArchiveFile(ctx context.Context, repoID string, opts ArchiveOptions, tmpPath string, desc *true_git.ArchiveDescriptor) (*ArchiveFile, error) {
	if lock, err := manager.lockGC(ctx, true); err != nil {
		return nil, err
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	if _, lock, err := werf.AcquireHostLock(ctx, fmt.Sprintf("git_archive.%s_%s", repoID, util.ObjectToHashKey(opts)), lockgate.AcquireOptions{}); err != nil {
		return nil, err
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	if archiveFile, err := manager.GetArchiveFile(ctx, repoID, opts); err != nil {
		return nil, err
	} else if archiveFile != nil {
		return archiveFile, nil
	}

	if err := os.MkdirAll(manager.ArchivesCacheDir, 0777); err != nil {
		return nil, fmt.Errorf("unable to create dir %q: %s", manager.ArchivesCacheDir, err)
	}

	metadataPath := filepath.Join(manager.ArchivesCacheDir, archiveMetadataFileName(repoID, opts))
	if metadata, err := json.Marshal(desc); err != nil {
		return nil, fmt.Errorf("error marshalling archive %s %s metadata json: %s", repoID, opts.Commit, err)
	} else {
		if err := ioutil.WriteFile(metadataPath, metadata, 0644); err != nil {
			return nil, fmt.Errorf("error writing %s: %s", metadataPath, err)
		}
	}

	path := filepath.Join(manager.ArchivesCacheDir, archiveFileName(repoID, opts))

	if err := os.Rename(tmpPath, path); err != nil {
		return nil, fmt.Errorf("unable to rename %s to %s: %s", tmpPath, path, err)
	}

	return &ArchiveFile{FilePath: path, Descriptor: desc}, nil
}

func (manager *GitDataManager) getPatchesCacheFilePath(repoID, fromCommit, toCommit string) string {
	return filepath.Join(manager.PatchesCacheDir, repoID, fmt.Sprintf("%s_%s.patch", fromCommit, toCommit))
}

func (manager *GitDataManager) GetPatchFile(ctx context.Context, repoID string, opts PatchOptions) (*PatchFile, error) {
	if lock, err := manager.lockGC(ctx, true); err != nil {
		return nil, err
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	metadataPath := filepath.Join(manager.PatchesCacheDir, patchMetadataFileName(repoID, opts))
	if exists, err := util.FileExists(metadataPath); err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}

	path := filepath.Join(manager.PatchesCacheDir, patchFileName(repoID, opts))
	if exists, err := util.FileExists(path); err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}

	if data, err := ioutil.ReadFile(metadataPath); err != nil {
		return nil, fmt.Errorf("error reading %s: %s", metadataPath, err)
	} else {
		var desc *true_git.PatchDescriptor
		if err := json.Unmarshal(data, &desc); err != nil {
			return nil, fmt.Errorf("error unmarshalling json from %s: %s", metadataPath, err)
		}

		return &PatchFile{FilePath: path, Descriptor: desc}, nil
	}
}

func (manager *GitDataManager) CreatePatchFile(ctx context.Context, repoID string, opts PatchOptions, tmpPath string, desc *true_git.PatchDescriptor) (*PatchFile, error) {
	if lock, err := manager.lockGC(ctx, true); err != nil {
		return nil, err
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	if _, lock, err := werf.AcquireHostLock(ctx, fmt.Sprintf("git_patch.%s_%s", repoID, util.ObjectToHashKey(opts)), lockgate.AcquireOptions{}); err != nil {
		return nil, err
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	if patchFile, err := manager.GetPatchFile(ctx, repoID, opts); err != nil {
		return nil, err
	} else if patchFile != nil {
		return patchFile, nil
	}

	if err := os.MkdirAll(manager.PatchesCacheDir, 0777); err != nil {
		return nil, fmt.Errorf("unable to create dir %q: %s", manager.PatchesCacheDir, err)
	}

	metadataPath := filepath.Join(manager.PatchesCacheDir, patchMetadataFileName(repoID, opts))
	if metadata, err := json.Marshal(desc); err != nil {
		return nil, fmt.Errorf("error marshalling patch %s %s %s metadata json: %s", repoID, opts.FromCommit, opts.ToCommit, err)
	} else {
		if err := ioutil.WriteFile(metadataPath, append(metadata, '\n'), 0644); err != nil {
			return nil, fmt.Errorf("error writing %s: %s", metadataPath, err)
		}
	}

	path := filepath.Join(manager.PatchesCacheDir, patchFileName(repoID, opts))

	if err := os.Rename(tmpPath, path); err != nil {
		return nil, fmt.Errorf("unable to rename %s to %s: %s", tmpPath, path, err)
	}

	return &PatchFile{FilePath: path, Descriptor: desc}, nil
}

func patchMetadataFileName(repoID string, opts PatchOptions) string {
	return fmt.Sprintf("%s_%s.meta.json", repoID, util.ObjectToHashKey(opts))
}

func patchFileName(repoID string, opts PatchOptions) string {
	return fmt.Sprintf("%s_%s.patch", repoID, util.ObjectToHashKey(opts))
}

func archiveMetadataFileName(repoID string, opts ArchiveOptions) string {
	return fmt.Sprintf("%s_%s.meta.json", repoID, util.ObjectToHashKey(opts))
}

func archiveFileName(repoID string, opts ArchiveOptions) string {
	return fmt.Sprintf("%s_%s.tar", repoID, util.ObjectToHashKey(opts))
}
