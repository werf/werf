package image

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/werf/lockgate"

	"github.com/werf/werf/pkg/werf"

	"github.com/werf/werf/pkg/util"
)

const (
	ManifestCacheVersion = "1"
)

type ManifestCache struct {
	CacheDir string
}

type ManifestCacheRecord struct {
	AccessTimestamp int64
	Info            *Info
}

func NewManifestCache(cacheDir string) *ManifestCache {
	return &ManifestCache{CacheDir: cacheDir}
}

func (cache *ManifestCache) GetImageInfo(imageName string) (*Info, error) {
	if lock, err := cache.lock(imageName); err != nil {
		return nil, err
	} else {
		defer cache.unlock(lock)
	}

	now := time.Now()

	if record, err := cache.readRecord(imageName); err != nil {
		return nil, err
	} else if record != nil {
		record.AccessTimestamp = now.Unix()
		if err := cache.writeRecord(record); err != nil {
			return nil, err
		}
		return record.Info, nil
	} else {
		return nil, nil
	}
}

func (cache *ManifestCache) StoreImageInfo(imgInfo *Info) error {
	if lock, err := cache.lock(imgInfo.Name); err != nil {
		return err
	} else {
		defer cache.unlock(lock)
	}

	record := &ManifestCacheRecord{
		AccessTimestamp: time.Now().Unix(),
		Info:            imgInfo,
	}
	return cache.writeRecord(record)
}

func (cache *ManifestCache) readRecord(imageName string) (*ManifestCacheRecord, error) {
	filePath := cache.constructFilePathForImage(imageName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("error accessing %s: %s", filePath, err)
	}

	if dataBytes, err := ioutil.ReadFile(filePath); err != nil {
		return nil, fmt.Errorf("error reading %s: %s", filePath, err)
	} else {
		record := &ManifestCacheRecord{}
		if err := json.Unmarshal(dataBytes, record); err != nil {
			return nil, fmt.Errorf("error unmarshalling json from %s: %s", filePath, err)
		}
		return record, nil
	}
}

func (cache *ManifestCache) writeRecord(record *ManifestCacheRecord) error {
	filePath := cache.constructFilePathForImage(record.Info.Name)

	dirPath := filepath.Dir(filePath)
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return fmt.Errorf("error creating dir %s: %s", dirPath, err)
	}

	if dataBytes, err := json.Marshal(record); err != nil {
		return fmt.Errorf("error marshalling json: %s", err)
	} else {
		if err := ioutil.WriteFile(filePath, append(dataBytes, []byte("\n")...), 0644); err != nil {
			return fmt.Errorf("error writing %s: %s", filePath, err)
		}
		return nil
	}
}

func (cache *ManifestCache) constructFilePathForImage(imageName string) string {
	return filepath.Join(cache.CacheDir, util.Sha256Hash(imageName))
}

func (cache *ManifestCache) lock(imageName string) (lockgate.LockHandle, error) {
	lockName := fmt.Sprintf("manifest_cache.%s", imageName)
	if _, lock, err := werf.AcquireHostLock(lockName, lockgate.AcquireOptions{}); err != nil {
		return lockgate.LockHandle{}, fmt.Errorf("cannot acquire %s host lock: %s", lockName, err)
	} else {
		return lock, nil
	}
}

func (cache *ManifestCache) unlock(lock lockgate.LockHandle) error {
	if err := werf.ReleaseHostLock(lock); err != nil {
		return fmt.Errorf("cannot release %s host lock: %s", lock.LockName, err)
	}
	return nil
}
