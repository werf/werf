package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/flant/logboek"

	"github.com/flant/lockgate"

	"github.com/werf/werf/pkg/werf"

	"k8s.io/apimachinery/pkg/util/json"

	"github.com/werf/werf/pkg/image"
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
	return fmt.Sprintf("%s", cache.CacheDir)
}

func (cache *FileStagesStorageCache) invalidateIfOldCacheExists(projectName string) error {
	if lock, err := cache.lock(); err != nil {
		return err
	} else {
		defer cache.unlock(lock)
	}

	oldCacheDir := filepath.Join(werf.GetLocalCacheDir(), "stages_storage_01")
	if _, lock, err := werf.AcquireHostLock(oldCacheDir, lockgate.AcquireOptions{}); err != nil {
		return err
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	currentProjectCacheDir := filepath.Join(cache.CacheDir, projectName)
	oldProjectCacheDir := filepath.Join(oldCacheDir, projectName)

	if _, err := os.Stat(oldProjectCacheDir); os.IsNotExist(err) {
		// ok
		return nil
	} else if err != nil {
		return fmt.Errorf("error accessing %s: %s", oldProjectCacheDir, err)
	} else {
		// remove old cache and new cache as well
		logboek.Default.LogF("Removing current stages storage project cache dir: %s\n", currentProjectCacheDir)
		if err := os.RemoveAll(currentProjectCacheDir); err != nil {
			return fmt.Errorf("error removing %s: %s", currentProjectCacheDir, err)
		}
		logboek.Default.LogF("Removing old stages storage project cache dir: %s\n", oldProjectCacheDir)
		if err := os.RemoveAll(oldProjectCacheDir); err != nil {
			return fmt.Errorf("error removing %s: %s", oldProjectCacheDir, err)
		}

		return nil
	}
}

func (cache *FileStagesStorageCache) GetAllStages(projectName string) (bool, []image.StageID, error) {
	if err := cache.invalidateIfOldCacheExists(projectName); err != nil {
		return false, nil, err
	}

	sigDir := filepath.Join(cache.CacheDir, projectName)

	if _, err := os.Stat(sigDir); os.IsNotExist(err) {
		return false, nil, nil
	} else if err != nil {
		return false, nil, fmt.Errorf("error accessing %s: %s", sigDir, err)
	}

	var res []image.StageID

	if entries, err := ioutil.ReadDir(sigDir); err != nil {
		return false, nil, fmt.Errorf("error reading directory %s files: %s", sigDir, err)
	} else {
		for _, finfo := range entries {
			if _, stages, err := cache.GetStagesBySignature(projectName, finfo.Name()); err != nil {
				return false, nil, err
			} else {
				res = append(res, stages...)
			}
		}
	}

	return true, res, nil
}

func (cache *FileStagesStorageCache) DeleteAllStages(projectName string) error {
	projectCacheDir := filepath.Join(cache.CacheDir, projectName)
	if err := os.RemoveAll(projectCacheDir); err != nil {
		return fmt.Errorf("unable to remove %s: %s", projectCacheDir, err)
	}
	return nil
}

func (cache *FileStagesStorageCache) GetStagesBySignature(projectName, signature string) (bool, []image.StageID, error) {
	if err := cache.invalidateIfOldCacheExists(projectName); err != nil {
		return false, nil, err
	}

	sigFile := filepath.Join(cache.CacheDir, projectName, signature)

	if _, err := os.Stat(sigFile); os.IsNotExist(err) {
		return false, nil, nil
	} else if err != nil {
		logboek.ErrF("Error accessing file %s: %s: will ignore cache\n", sigFile, err)
		return false, nil, nil
	}

	dataBytes, err := ioutil.ReadFile(sigFile)
	if err != nil {
		logboek.ErrF("Error reading file %s: %s: will ignore cache\n", sigFile, err)
		return false, nil, nil
	}

	res := &StagesStorageCacheRecord{}
	if err := json.Unmarshal(dataBytes, res); err != nil {
		logboek.ErrF("Error unmarshalling json from %s: %s: will ignore cache\n", sigFile, err)
		return false, nil, nil
	}

	return true, res.Stages, nil
}

func (cache *FileStagesStorageCache) StoreStagesBySignature(projectName, signature string, stages []image.StageID) error {
	if err := cache.invalidateIfOldCacheExists(projectName); err != nil {
		return err
	}

	if lock, err := cache.lock(); err != nil {
		return err
	} else {
		defer cache.unlock(lock)
	}

	sigDir := filepath.Join(cache.CacheDir, projectName)
	sigFile := filepath.Join(sigDir, signature)
	if err := os.MkdirAll(sigDir, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create dir %s: %s", sigDir, err)
	}

	dataBytes, err := json.Marshal(StagesStorageCacheRecord{Stages: stages})
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(sigFile, append(dataBytes, []byte("\n")...), 0644); err != nil {
		return fmt.Errorf("error writing file %s: %s", sigFile, err)
	}

	return nil
}

func (cache *FileStagesStorageCache) DeleteStagesBySignature(projectName, signature string) error {
	if err := cache.invalidateIfOldCacheExists(projectName); err != nil {
		return err
	}

	if lock, err := cache.lock(); err != nil {
		return err
	} else {
		defer cache.unlock(lock)
	}

	sigDir := filepath.Join(cache.CacheDir, projectName)
	sigFile := filepath.Join(sigDir, signature)

	if err := os.RemoveAll(sigFile); err != nil {
		return fmt.Errorf("error removing %s: %s", sigFile, err)
	}
	return nil
}

func (cache *FileStagesStorageCache) lock() (lockgate.LockHandle, error) {
	_, lock, err := werf.AcquireHostLock(cache.CacheDir, lockgate.AcquireOptions{})
	return lock, err
}

func (cache *FileStagesStorageCache) unlock(lock lockgate.LockHandle) error {
	return werf.ReleaseHostLock(lock)
}
