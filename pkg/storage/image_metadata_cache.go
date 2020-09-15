package storage

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

var imageMetadataCache *ImageMetadataCache

const (
	ImageMetadataCacheVersion = "1"
)

type ImageMetadataCache struct {
	CacheDir string
}

type ImageMetadataCacheRecord struct {
	AccessTimestamp int64
	ImageMetadata   *ImageMetadata
}

func GetImageMetadataCache() *ImageMetadataCache {
	if imageMetadataCache == nil {
		imageMetadataCache = &ImageMetadataCache{
			CacheDir: filepath.Join(werf.GetLocalCacheDir(), "image_metadata", ImageMetadataCacheVersion),
		}
	}

	return imageMetadataCache
}

func (cache *ImageMetadataCache) GetImageMetadata(ctx context.Context, storageName, imageName, commit string) (*ImageMetadata, error) {
	logProcess := logboek.Context(ctx).Debug().LogProcess("-- ImageMetadataCache.GetImageMetadata %s %s %s", storageName, imageName, commit)
	logProcess.Start()
	defer logProcess.End()

	if lock, err := cache.lock(ctx, storageName, imageName, commit); err != nil {
		return nil, err
	} else {
		defer cache.unlock(lock)
	}

	now := time.Now()
	if record, err := cache.readRecord(storageName, imageName, commit); err != nil {
		return nil, err
	} else if record != nil {
		record.AccessTimestamp = now.Unix()
		if err := cache.writeRecord(storageName, imageName, commit, record); err != nil {
			return nil, err
		}
		return record.ImageMetadata, nil
	} else {
		return nil, nil
	}
}

func (cache *ImageMetadataCache) StoreImageMetadata(ctx context.Context, storageName, imageName, commit string, imageMetadata *ImageMetadata) error {
	logProcess := logboek.Context(ctx).Debug().LogProcess("-- ImageMetadataCache.StoreImageMetadata %s %s %s", storageName, imageName, commit)
	logProcess.Start()
	defer logProcess.End()

	if lock, err := cache.lock(ctx, storageName, imageName, commit); err != nil {
		return err
	} else {
		defer cache.unlock(lock)
	}

	record := &ImageMetadataCacheRecord{
		AccessTimestamp: time.Now().Unix(),
		ImageMetadata:   imageMetadata,
	}
	return cache.writeRecord(storageName, imageName, commit, record)
}

func (cache *ImageMetadataCache) readRecord(storageName, imageName, commit string) (*ImageMetadataCacheRecord, error) {
	filePath := cache.constructFilePathForImage(storageName, imageName, commit)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("error accessing %s: %s", filePath, err)
	}

	if dataBytes, err := ioutil.ReadFile(filePath); err != nil {
		return nil, fmt.Errorf("error reading %s: %s", filePath, err)
	} else {
		record := &ImageMetadataCacheRecord{}
		if err := json.Unmarshal(dataBytes, record); err != nil {
			return nil, fmt.Errorf("error unmarshalling json from %s: %s", filePath, err)
		}
		return record, nil
	}
}

func (cache *ImageMetadataCache) writeRecord(storageName, imageName, commit string, record *ImageMetadataCacheRecord) error {
	filePath := cache.constructFilePathForImage(storageName, imageName, commit)

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

func (cache *ImageMetadataCache) constructFilePathForImage(storageName, imageName, commit string) string {
	return filepath.Join(cache.CacheDir, slug.Slug(storageName), util.Sha256Hash(imageName, commit))
}

func (cache *ImageMetadataCache) lock(ctx context.Context, storageName, imageName, commit string) (lockgate.LockHandle, error) {
	lockName := fmt.Sprintf("image_metadata_cache.%s.%s.%s", slug.Slug(storageName), imageName, commit)
	if _, lock, err := werf.AcquireHostLock(ctx, lockName, lockgate.AcquireOptions{}); err != nil {
		return lockgate.LockHandle{}, fmt.Errorf("cannot acquire %s host lock: %s", lockName, err)
	} else {
		return lock, nil
	}
}

func (cache *ImageMetadataCache) unlock(lock lockgate.LockHandle) error {
	if err := werf.ReleaseHostLock(lock); err != nil {
		return fmt.Errorf("cannot release %s host lock: %s", lock.LockName, err)
	}
	return nil
}
