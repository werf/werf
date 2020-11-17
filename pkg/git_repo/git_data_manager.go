package git_repo

import (
	"fmt"
	"os"
	"path/filepath"

	uuid "github.com/satori/go.uuid"
	"github.com/werf/werf/pkg/werf"
)

const (
	GitDataCacheVersion = "1"
)

type GitDataManager struct {
	ArchivesCacheDir string
	PatchesCacheDir  string
	TmpDir           string
}

func NewCommonGitDataManager() *GitDataManager {
	return NewGitDataManager(
		filepath.Join(werf.GetLocalCacheDir(), "git_data", GitDataCacheVersion, "archives"),
		filepath.Join(werf.GetLocalCacheDir(), "git_data", GitDataCacheVersion, "patches"),
		filepath.Join(werf.GetLocalCacheDir(), "git_data", GitDataCacheVersion, "tmp"),
	)
}

func NewGitDataManager(archivesCacheDir, patchesCacheDir, tmpDir string) *GitDataManager {
	return &GitDataManager{ArchivesCacheDir: archivesCacheDir, PatchesCacheDir: patchesCacheDir, TmpDir: tmpDir}
}

func (manager *GitDataManager) GC() error {
	return nil
}

func (manager *GitDataManager) getArchiveCacheFilePath(repoID, commit string) string {
	return filepath.Join(manager.ArchivesCacheDir, repoID, fmt.Sprintf("%s.tar", commit))
}

func (manager *GitDataManager) NewTmpArchiveFile() (*ArchiveFile, error) {
	path := filepath.Join(manager.TmpDir, fmt.Sprintf("%s.tar", uuid.NewV4().String()))
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return nil, fmt.Errorf("unable to create dir %q: %s", filepath.Dir(path), err)
	}
	return &ArchiveFile{FilePath: path}, nil
}

func (manager *GitDataManager) GetArchiveFile(repoID, commit string) (*ArchiveFile, error) {
	path := manager.getArchiveCacheFilePath(repoID, commit)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("stat file %q failed: %s", path, err)
	}
	return &ArchiveFile{FilePath: path}, nil
}

func (manager *GitDataManager) PutArchiveFile(repoID, commit string, archiveFile *ArchiveFile) error {
	path := manager.getArchiveCacheFilePath(repoID, commit)
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return fmt.Errorf("unable to create dir %q: %s", filepath.Dir(path), err)
	}
	// TODO: put into cache, use lock for gc and another puts
	return nil
}

func (manager *GitDataManager) getPatchesCacheFilePath(repoID, fromCommit, toCommit string) string {
	return filepath.Join(manager.PatchesCacheDir, repoID, fmt.Sprintf("%s_%s.patch", fromCommit, toCommit))
}

func (manager *GitDataManager) NewTmpPatchFile() (*PatchFile, error) {
	path := filepath.Join(manager.TmpDir, fmt.Sprintf("%s.patch", uuid.NewV4().String()))
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return nil, fmt.Errorf("unable to create dir %q: %s", filepath.Dir(path), err)
	}

	return &PatchFile{FilePath: path}, nil
}

func (manager *GitDataManager) GetPatchFile(repoID, fromCommit, toCommit string) (*PatchFile, error) {
	path := manager.getPatchesCacheFilePath(repoID, fromCommit, toCommit)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("stat file %q failed: %s", path, err)
	}
	return &PatchFile{FilePath: path}, nil
}

func (manager *GitDataManager) PutPatchFile(repoID, fromCommit, toCommit string, patchFile *PatchFile) error {
	path := manager.getPatchesCacheFilePath(repoID, fromCommit, toCommit)
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return fmt.Errorf("unable to create dir %q: %s", filepath.Dir(path), err)
	}
	// TODO: put into cache, use lock for gc and another puts
	return nil
}
