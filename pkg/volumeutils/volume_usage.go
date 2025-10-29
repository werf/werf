package volumeutils

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/werf/werf/v2/pkg/third_party/minio/disk"
)

type VolumeUsage struct {
	UsedBytes  uint64
	TotalBytes uint64
}

func (vu VolumeUsage) Percentage() float64 {
	return (float64(vu.UsedBytes) / float64(vu.TotalBytes)) * 100
}

func (vu VolumeUsage) BytesToFree(targetVolumeUsagePercentage float64) uint64 {
	diffPercentage := vu.Percentage() - targetVolumeUsagePercentage
	allowedVolumeUsageToFree := math.Max(diffPercentage, 0)
	bytesToFree := uint64((float64(vu.TotalBytes) / 100.0) * allowedVolumeUsageToFree)
	return bytesToFree
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
