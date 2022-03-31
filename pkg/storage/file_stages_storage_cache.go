package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/util/json"

	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/werf"
)

type FileStagesStorageCache struct {
	CacheDir string
}

type StagesStorageCacheRecord struct {
	Stages []image.StageID `json:"stages"`
}

func NewFileStagesStorageCache(cacheDir string) *FileStagesStorageCache {
	return &FileStagesStorageCache{CacheDir: cacheDir}
}

func (cache *FileStagesStorageCache) String() string {
	return cache.CacheDir
}

func (cache *FileStagesStorageCache) GetAllStages(ctx context.Context, projectName string) (bool, []image.StageID, error) {
	sigDir := filepath.Join(cache.CacheDir, projectName)

	if _, err := os.Stat(sigDir); os.IsNotExist(err) {
		return false, nil, nil
	} else if err != nil {
		return false, nil, fmt.Errorf("error accessing %s: %w", sigDir, err)
	}

	var res []image.StageID

	if entries, err := ioutil.ReadDir(sigDir); err != nil {
		return false, nil, fmt.Errorf("error reading directory %s files: %w", sigDir, err)
	} else {
		for _, finfo := range entries {
			if _, stages, err := cache.GetStagesByDigest(ctx, projectName, finfo.Name()); err != nil {
				return false, nil, err
			} else {
				res = append(res, stages...)
			}
		}
	}

	return true, res, nil
}

func (cache *FileStagesStorageCache) DeleteAllStages(_ context.Context, projectName string) error {
	projectCacheDir := filepath.Join(cache.CacheDir, projectName)
	if err := os.RemoveAll(projectCacheDir); err != nil {
		return fmt.Errorf("unable to remove %s: %w", projectCacheDir, err)
	}
	return nil
}

func (cache *FileStagesStorageCache) GetStagesByDigest(ctx context.Context, projectName, digest string) (bool, []image.StageID, error) {
	sigFile := filepath.Join(cache.CacheDir, projectName, digest)

	if _, err := os.Stat(sigFile); os.IsNotExist(err) {
		return false, nil, nil
	} else if err != nil {
		logboek.Context(ctx).Error().LogF("Error accessing file %s: %s: will ignore cache\n", sigFile, err)
		return false, nil, nil
	}

	dataBytes, err := ioutil.ReadFile(sigFile)
	if err != nil {
		logboek.Context(ctx).Error().LogF("Error reading file %s: %s: will ignore cache\n", sigFile, err)
		return false, nil, nil
	}

	res := &StagesStorageCacheRecord{}
	if err := json.Unmarshal(dataBytes, res); err != nil {
		logboek.Context(ctx).Error().LogF("Error unmarshalling json from %s: %s: will ignore cache\n", sigFile, err)
		return false, nil, nil
	}

	return true, res.Stages, nil
}

func (cache *FileStagesStorageCache) StoreStagesByDigest(ctx context.Context, projectName, digest string, stages []image.StageID) error {
	if lock, err := cache.lock(ctx); err != nil {
		return err
	} else {
		defer cache.unlock(lock)
	}

	sigDir := filepath.Join(cache.CacheDir, projectName)
	sigFile := filepath.Join(sigDir, digest)
	if err := os.MkdirAll(sigDir, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create dir %s: %w", sigDir, err)
	}

	dataBytes, err := json.Marshal(StagesStorageCacheRecord{Stages: stages})
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(sigFile, append(dataBytes, []byte("\n")...), 0o644); err != nil {
		return fmt.Errorf("error writing file %s: %w", sigFile, err)
	}

	return nil
}

func (cache *FileStagesStorageCache) DeleteStagesByDigest(ctx context.Context, projectName, digest string) error {
	if lock, err := cache.lock(ctx); err != nil {
		return err
	} else {
		defer cache.unlock(lock)
	}

	sigDir := filepath.Join(cache.CacheDir, projectName)
	sigFile := filepath.Join(sigDir, digest)

	if err := os.RemoveAll(sigFile); err != nil {
		return fmt.Errorf("error removing %s: %w", sigFile, err)
	}
	return nil
}

func (cache *FileStagesStorageCache) lock(ctx context.Context) (lockgate.LockHandle, error) {
	_, lock, err := werf.AcquireHostLock(ctx, cache.CacheDir, lockgate.AcquireOptions{})
	return lock, err
}

func (cache *FileStagesStorageCache) unlock(lock lockgate.LockHandle) error {
	return werf.ReleaseHostLock(lock)
}
