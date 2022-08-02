package lrumeta

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
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

const LRUImagesCacheVersion = "1"

var CommonLRUImagesCache *LRUImagesCache

func Init() error {
	CommonLRUImagesCache = NewLRUImagesCache(filepath.Join(werf.GetLocalCacheDir(), "lru_images", LRUImagesCacheVersion))
	return nil
}

type LRUImagesCache struct {
	CacheDir string
}

type LRUImagesCacheRecord struct {
	AccessTimestampNanosec int64
	ImageRef               string
}

func NewLRUImagesCache(cacheDir string) *LRUImagesCache {
	return &LRUImagesCache{CacheDir: cacheDir}
}

func (cache *LRUImagesCache) AccessImage(ctx context.Context, imageRef string) error {
	logProcess := logboek.Context(ctx).Debug().LogProcess("-- LRUImagesCache.AccessImage %s", imageRef)
	logProcess.Start()
	defer logProcess.End()

	if lock, err := cache.lock(ctx, imageRef); err != nil {
		return err
	} else {
		defer cache.unlock(lock)
	}

	record := &LRUImagesCacheRecord{
		AccessTimestampNanosec: time.Now().UnixNano(),
		ImageRef:               imageRef,
	}

	return cache.writeRecord(record)
}

func (cache *LRUImagesCache) GetImageLastAccessTime(ctx context.Context, imageRef string) (time.Time, error) {
	logProcess := logboek.Context(ctx).Debug().LogProcess("-- LRUImagesCache.GetImageLastAccessTimestamp %s", imageRef)
	logProcess.Start()
	defer logProcess.End()

	if lock, err := cache.lock(ctx, imageRef); err != nil {
		return time.Time{}, err
	} else {
		defer cache.unlock(lock)
	}

	record, err := cache.readRecord(ctx, imageRef)
	if err != nil {
		return time.Time{}, err
	}

	if record == nil {
		return time.Time{}, nil
	}

	return time.Unix(record.AccessTimestampNanosec/1_000_000_000, record.AccessTimestampNanosec%1_000_000_000), nil
}

func (cache *LRUImagesCache) readRecord(ctx context.Context, imageRef string) (*LRUImagesCacheRecord, error) {
	filePath := cache.constructFilePathForImage(imageRef)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("error accessing %s: %w", filePath, err)
	}

	if dataBytes, err := ioutil.ReadFile(filePath); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", filePath, err)
	} else {
		record := &LRUImagesCacheRecord{}
		if err := json.Unmarshal(dataBytes, record); err != nil {
			logboek.Context(ctx).Error().LogF("WARNING: invalid lru images cache json record in file %s: %s: resetting record\n", filePath, err)
			return nil, nil
		}
		return record, nil
	}
}

func (cache *LRUImagesCache) writeRecord(record *LRUImagesCacheRecord) error {
	filePath := cache.constructFilePathForImage(record.ImageRef)

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

func (cache *LRUImagesCache) constructFilePathForImage(imageRef string) string {
	return filepath.Join(cache.CacheDir, util.Sha256Hash(imageRef))
}

func (cache *LRUImagesCache) lock(ctx context.Context, imageRef string) (lockgate.LockHandle, error) {
	lockName := fmt.Sprintf("lru_images_cache.%s", imageRef)
	if _, lock, err := werf.AcquireHostLock(ctx, lockName, lockgate.AcquireOptions{}); err != nil {
		return lockgate.LockHandle{}, fmt.Errorf("cannot acquire %s host lock: %w", lockName, err)
	} else {
		return lock, nil
	}
}

func (cache *LRUImagesCache) unlock(lock lockgate.LockHandle) error {
	if err := werf.ReleaseHostLock(lock); err != nil {
		return fmt.Errorf("cannot release %s host lock: %w", lock.LockName, err)
	}
	return nil
}
