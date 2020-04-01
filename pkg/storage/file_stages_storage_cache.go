package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/util/json"

	"github.com/flant/shluz"
	"github.com/flant/werf/pkg/image"
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

func (cache *FileStagesStorageCache) GetAllStages(projectName string) (bool, []image.StageID, error) {
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

func (cache *FileStagesStorageCache) GetStagesBySignature(projectName, signature string) (bool, []image.StageID, error) {
	sigFile := filepath.Join(cache.CacheDir, projectName, signature)

	if _, err := os.Stat(sigFile); os.IsNotExist(err) {
		return false, nil, nil
	} else if err != nil {
		return false, nil, fmt.Errorf("error accessing file %s: %s", sigFile, err)
	}

	dataBytes, err := ioutil.ReadFile(sigFile)
	if err != nil {
		return false, nil, fmt.Errorf("error reading file %s: %s", sigFile, err)
	}

	res := &StagesStorageCacheRecord{}
	if err := json.Unmarshal(dataBytes, res); err != nil {
		return false, nil, fmt.Errorf("error unmarshalling json from %s: %s", sigFile, err)
	}

	return true, res.Stages, nil
}

func (cache *FileStagesStorageCache) StoreStagesBySignature(projectName, signature string, stages []image.StageID) error {
	if err := cache.lock(); err != nil {
		return err
	}
	defer cache.unlock()

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
	if err := cache.lock(); err != nil {
		return err
	}
	defer cache.unlock()

	sigDir := filepath.Join(cache.CacheDir, projectName)
	sigFile := filepath.Join(sigDir, signature)

	if err := os.RemoveAll(sigFile); err != nil {
		return fmt.Errorf("error removing %s: %s", sigFile, err)
	}
	return nil
}

func (cache *FileStagesStorageCache) lock() error {
	// NOTE maybe shluz is an overkill for this kind of locks
	if err := shluz.Lock(cache.CacheDir, shluz.LockOptions{}); err != nil {
		return fmt.Errorf("shluz lock %s failed: %s", cache.CacheDir, err)
	}
	return nil
}

func (cache *FileStagesStorageCache) unlock() error {
	return shluz.Unlock(cache.CacheDir)
}
