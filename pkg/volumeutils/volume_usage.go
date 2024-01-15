package volumeutils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/werf/werf/pkg/third_party/minio/disk"
)

type VolumeUsage struct {
	UsedBytes  uint64
	TotalBytes uint64
	Percentage float64
}

func GetVolumeUsageByPath(ctx context.Context, path string) (VolumeUsage, error) {
	di, err := disk.GetInfo(path, true)
	if err != nil {
		return VolumeUsage{}, fmt.Errorf("unable to get disk info: %w", err)
	}

	usedBytes := di.Total - di.Free
	return VolumeUsage{
		UsedBytes:  usedBytes,
		TotalBytes: di.Total,
		Percentage: (float64(usedBytes) / float64(di.Total)) * 100,
	}, nil
}

func DirSizeBytes(path string) (uint64, error) {
	var size uint64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing %q: %w", path, err)
		}
		if !info.IsDir() {
			size += uint64(info.Size())
		}
		return err
	})
	return size, err
}
