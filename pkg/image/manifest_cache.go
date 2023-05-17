package image

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/slug"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

const (
	ManifestCacheVersion = "5"
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

func (cache *ManifestCache) GetImageInfo(ctx context.Context, storageName, imageName string) (*Info, error) {
	logProcess := logboek.Context(ctx).Debug().LogProcess("-- ManifestCache.GetImageInfo %s %s", storageName, imageName)
	logProcess.Start()
	defer logProcess.End()

	if lock, err := cache.lock(ctx, storageName, imageName); err != nil {
		return nil, err
	} else {
		defer cache.unlock(lock)
	}

	now := time.Now()

	record, err := cache.readRecord(ctx, storageName, imageName)
	switch {
	case err != nil:
		return nil, err
	case record != nil:
		record.AccessTimestamp = now.Unix()
		if err := cache.writeRecord(storageName, record); err != nil {
			return nil, err
		}

		return record.Info, nil
	default:
		return nil, nil
	}
}

func (cache *ManifestCache) StoreImageInfo(ctx context.Context, storageName string, imgInfo *Info) error {
	logProcess := logboek.Context(ctx).Debug().LogProcess("-- ManifestCache.StoreImageInfo %s %s", storageName, imgInfo.Name)
	logProcess.Start()
	defer logProcess.End()

	if lock, err := cache.lock(ctx, storageName, imgInfo.Name); err != nil {
		return err
	} else {
		defer cache.unlock(lock)
	}

	record := &ManifestCacheRecord{
		AccessTimestamp: time.Now().Unix(),
		Info:            imgInfo,
	}
	return cache.writeRecord(storageName, record)
}

func (cache *ManifestCache) readRecord(ctx context.Context, storageName, imageName string) (*ManifestCacheRecord, error) {
	filePath := cache.constructFilePathForImage(storageName, imageName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("error accessing %s: %w", filePath, err)
	}

	if dataBytes, err := ioutil.ReadFile(filePath); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", filePath, err)
	} else {
		record := &ManifestCacheRecord{}
		if err := json.Unmarshal(dataBytes, record); err != nil {
			logboek.Context(ctx).Error().LogF("WARNING: invalid manifests cache json record in file %s: %s: resetting record\n", filePath, err)
			return nil, nil
		}
		return record, nil
	}
}

func (cache *ManifestCache) writeRecord(storageName string, record *ManifestCacheRecord) error {
	filePath := cache.constructFilePathForImage(storageName, record.Info.Name)

	dirPath := filepath.Dir(filePath)
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return fmt.Errorf("error creating dir %s: %w", dirPath, err)
	}

	if dataBytes, err := json.Marshal(record); err != nil {
		return fmt.Errorf("error marshalling json: %w", err)
	} else {
		if err := ioutil.WriteFile(filePath, append(dataBytes, []byte("\n")...), 0o644); err != nil {
			return fmt.Errorf("error writing %s: %w", filePath, err)
		}
		return nil
	}
}

func (cache *ManifestCache) constructFilePathForImage(storageName, imageName string) string {
	return filepath.Join(cache.CacheDir, slug.Slug(storageName), util.Sha256Hash(imageName))
}

func (cache *ManifestCache) lock(ctx context.Context, storageName, imageName string) (lockgate.LockHandle, error) {
	lockName := fmt.Sprintf("manifest_cache.%s.%s", slug.Slug(storageName), imageName)
	if _, lock, err := werf.AcquireHostLock(ctx, lockName, lockgate.AcquireOptions{}); err != nil {
		return lockgate.LockHandle{}, fmt.Errorf("cannot acquire %s host lock: %w", lockName, err)
	} else {
		return lock, nil
	}
}

func (cache *ManifestCache) unlock(lock lockgate.LockHandle) error {
	if err := werf.ReleaseHostLock(lock); err != nil {
		return fmt.Errorf("cannot release %s host lock: %w", lock.LockName, err)
	}
	return nil
}
