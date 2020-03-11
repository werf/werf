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

type ImageInfosCacheData struct {
	ImagesDescs []*image.Info `json:"imagesDescs"`
}

func NewFileStagesStorageCache(cacheDir string) *FileStagesStorageCache {
	return &FileStagesStorageCache{CacheDir: cacheDir}
}

func (cache *FileStagesStorageCache) GetImagesBySignature(projectName, signature string) (bool, []*image.Info, error) {
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

	res := &ImageInfosCacheData{}
	if err := json.Unmarshal(dataBytes, res); err != nil {
		return false, nil, fmt.Errorf("error unmarshalling json from %s: %s", sigFile, err)
	}

	return true, res.ImagesDescs, nil
}

func (cache *FileStagesStorageCache) StoreImagesBySignature(projectName, signature string, imagesDescs []*image.Info) error {
	if err := cache.lock(); err != nil {
		return err
	}
	defer cache.unlock()

	sigDir := filepath.Join(cache.CacheDir, projectName)
	sigFile := filepath.Join(sigDir, signature)
	if err := os.MkdirAll(sigDir, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create dir %s: %s", sigDir, err)
	}

	dataBytes, err := json.Marshal(ImageInfosCacheData{ImagesDescs: imagesDescs})
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(sigFile, append(dataBytes, []byte("\n")...), 0644); err != nil {
		return fmt.Errorf("error writing file %s: %s", sigFile, err)
	}

	return nil
}

func (cache *FileStagesStorageCache) lock() error {
	// TODO: maybe shluz is an overkill for this kind of locks
	if err := shluz.Lock(cache.CacheDir, shluz.LockOptions{}); err != nil {
		return fmt.Errorf("shluz lock %s failed: %s", cache.CacheDir, err)
	}
	return nil
}

func (cache *FileStagesStorageCache) unlock() error {
	return shluz.Unlock(cache.CacheDir)
}
