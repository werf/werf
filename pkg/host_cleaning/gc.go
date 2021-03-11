package host_cleaning

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"

	"github.com/dustin/go-humanize"
	"github.com/minio/minio/pkg/disk"

	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/storage/lrumeta"
	"github.com/werf/werf/pkg/werf"

	"github.com/werf/kubedog/pkg/utils"
)

const (
	DefaultAllowedVolumeUsagePercentageThreshold float64 = 80.0
	LockName                                             = "host_cleaning.GC"
)

// AcquireSharedStorageLock should be called in every process that normally work with the storage
func AcquireSharedHostStorageLock(ctx context.Context) (lockgate.LockHandle, error) {
	_, lock, err := werf.AcquireHostLock(ctx, LockName, lockgate.AcquireOptions{Shared: true})
	if err != nil {
		return lockgate.LockHandle{}, fmt.Errorf("unable to acquire host lock %q: %s", LockName, err)
	}
	return lock, nil
}

func GetLocalDockerServerStoragePath(ctx context.Context) (string, error) {
	dockerInfo, err := docker.Info(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to get docker info: %s", err)
	}

	var storagePath string

	if dockerInfo.OperatingSystem == "Docker Desktop" {
		switch runtime.GOOS {
		case "windows":
			storagePath = filepath.Join(os.Getenv("HOMEDRIVE"), `\\ProgramData\DockerDesktop\vm-data\`)

		case "darwin":
			storagePath = filepath.Join(os.Getenv("HOME"), "Library/Containers/com.docker.docker/Data")
		}
	} else {
		storagePath = dockerInfo.DockerRootDir
	}

	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("error accessing %q: %s", storagePath, err)
	}
	return storagePath, nil
}

type LocalDockerServerGCOptions struct {
	AllowedVolumeUsagePercentageThreshold int64
	DryRun                                bool
	Force                                 bool
	DockerServerStoragePath               string
}

func getAllowedVolumeUsagePercentageThreshold(opts LocalDockerServerGCOptions) float64 {
	var percentageThreshold float64
	if opts.AllowedVolumeUsagePercentageThreshold > 0 {
		percentageThreshold = float64(opts.AllowedVolumeUsagePercentageThreshold)
	} else {
		percentageThreshold = DefaultAllowedVolumeUsagePercentageThreshold
	}
	return percentageThreshold
}

func getDockerServerStoragePath(ctx context.Context, opts LocalDockerServerGCOptions) (string, error) {
	var dockerServerStoragePath string
	if opts.DockerServerStoragePath != "" {
		dockerServerStoragePath = opts.DockerServerStoragePath
	} else {
		path, err := GetLocalDockerServerStoragePath(ctx)
		if err != nil {
			return "", err
		}
		dockerServerStoragePath = path
	}

	return dockerServerStoragePath, nil
}

func ShouldRunGCForLocalDockerServer(ctx context.Context, opts LocalDockerServerGCOptions) (bool, error) {
	percentageThreshold := getAllowedVolumeUsagePercentageThreshold(opts)

	dockerServerStoragePath, err := getDockerServerStoragePath(ctx, opts)
	if err != nil {
		return false, fmt.Errorf("error getting local docker server storage path: %s", err)
	}

	if dockerServerStoragePath == "" {
		return false, nil
	}

	vu, err := GetVolumeUsageByPath(ctx, dockerServerStoragePath)
	if err != nil {
		return false, fmt.Errorf("error getting volume usage by path %q: %s", dockerServerStoragePath, err)
	}

	return vu.Percentage > percentageThreshold, nil
}

func RunGCForLocalDockerServer(ctx context.Context, opts LocalDockerServerGCOptions) error {
	if _, lock, err := werf.AcquireHostLock(ctx, LockName, lockgate.AcquireOptions{}); err != nil {
		return fmt.Errorf("unable to acquire host lock %q: %s", LockName, err)
	} else {
		defer werf.ReleaseHostLock(lock)
	}

	percentageThreshold := getAllowedVolumeUsagePercentageThreshold(opts)

	dockerServerStoragePath, err := getDockerServerStoragePath(ctx, opts)
	if err != nil {
		return fmt.Errorf("error getting local docker server storage path: %s", err)
	}

	if dockerServerStoragePath == "" {
		return nil
	}

	for {
		// Remove all images by old cache-version without lru? — No, not correct. These will be cleaned up by lru anyway.
		// TODO: Filter out local stages

		vu, err := GetVolumeUsageByPath(ctx, dockerServerStoragePath)
		if err != nil {
			return fmt.Errorf("error getting volume usage by path %q: %s", dockerServerStoragePath, err)
		}

		if vu.Percentage <= percentageThreshold {
			logboek.Context(ctx).Default().LogBlock("Local docker server storage check").Do(func() {
				logboek.Context(ctx).Default().LogF("Docker server storage path: %s\n", dockerServerStoragePath)
				logboek.Context(ctx).Default().LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
				logboek.Context(ctx).Default().LogF("Allowed usage percentage: %s <= %s — %s\n", utils.GreenF("%0.2f%%", vu.Percentage), utils.BlueF("%0.2f%%", percentageThreshold), utils.GreenF("OK"))
			})

			break
		}

		percentageToFree := vu.Percentage - percentageThreshold
		bytesToFree := uint64(float64(vu.TotalBytes) / 100.0 * percentageToFree)

		filterSet := filters.NewArgs()
		filterSet.Add("label", fmt.Sprintf("%s", image.WerfLabel))
		filterSet.Add("label", fmt.Sprintf("%s", image.WerfStageDigestLabel))
		images, err := docker.Images(ctx, types.ImageListOptions{Filters: filterSet})
		if err != nil {
			return fmt.Errorf("unable tio get docker images: %s", err)
		}

		var totalImagesBytes uint64
		var imagesDescs []*LocalImageDesc

		for _, imageSummary := range images {
			data, _ := json.Marshal(imageSummary)
			logboek.Context(ctx).Debug().LogF("Image summary:\n%s\n---\n", data)

			totalImagesBytes += uint64(imageSummary.Size)

			lastUsedAt := time.Unix(imageSummary.Created, 0)

			for _, ref := range imageSummary.RepoTags {
				if ref == "<none>:<none>" {
					continue
				}

				lastRecentlyUsedAt, err := lrumeta.CommonLRUImagesCache.GetImageLastAccessTime(ctx, ref)
				if err != nil {
					return fmt.Errorf("error accessing last recently used images cache: %s", err)
				}

				if lastRecentlyUsedAt.IsZero() {
					continue
				}

				lastUsedAt = lastRecentlyUsedAt
			}

			desc := &LocalImageDesc{
				ImageSummary: imageSummary,
				LastUsedAt:   lastUsedAt,
			}
			imagesDescs = append(imagesDescs, desc)
		}

		sort.Sort(ImagesLruSort(imagesDescs))

		logboek.Context(ctx).Default().LogBlock("Local docker server storage check").Do(func() {
			logboek.Context(ctx).Default().LogF("Docker server storage path: %s\n", dockerServerStoragePath)
			logboek.Context(ctx).Default().LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
			logboek.Context(ctx).Default().LogF("Allowed percentage level exceeded: %s > %s — %s\n", utils.RedF("%0.2f%%", vu.Percentage), utils.BlueF("%0.2f%%", percentageThreshold), utils.RedF("HIGH DISK USAGE"))
			logboek.Context(ctx).Default().LogF("Needed to free: %s\n", utils.RedF("%s", humanize.Bytes(bytesToFree)))
			logboek.Context(ctx).Default().LogF("Available images to free: %s\n", utils.YellowF("%d (~ %s)", len(imagesDescs), humanize.Bytes(totalImagesBytes)))
		})

		var freedBytes uint64
		var freedImagesCount uint64

		if len(imagesDescs) > 0 {
			if err := logboek.Context(ctx).Default().LogProcess("Running cleanup for least recently used docker images created by werf").DoError(func() error {
				for _, desc := range imagesDescs {
					imageRemovalFailed := false
					for _, ref := range desc.ImageSummary.RepoTags {
						var args []string

						if ref == "<none>:<none>" {
							args = append(args, desc.ImageSummary.ID)
						} else {
							args = append(args, ref)
						}

						if opts.Force {
							args = append(args, "--force")
						}

						logboek.Context(ctx).Default().LogF("Removing %s\n", ref)
						if opts.DryRun {
							continue
						}

						if err := docker.CliRmi(ctx, args...); err != nil {
							logboek.Context(ctx).Warn().LogF("failed to remove local docker image %q: %s\n", ref, err)
							imageRemovalFailed = true
						}
					}

					if !imageRemovalFailed {
						freedBytes += uint64(desc.ImageSummary.Size)
						freedImagesCount++
					}

					if freedBytes > bytesToFree {
						break
					}
				}

				logboek.Context(ctx).Default().LogF("Freed images: %s\n", utils.GreenF("%d (~ %s)", freedImagesCount, humanize.Bytes(freedBytes)))

				return nil
			}); err != nil {
				return err
			}
		} else {
			logboek.Context(ctx).Warn().LogF("WARNING: Detected high docker storage volume usage, while no werf images available to cleanup!\n")
			logboek.Context(ctx).Warn().LogOptionalLn()
		}

		commonOptions := CommonOptions{
			RmContainersThatUseWerfImages: opts.Force,
			SkipUsedImages:                !opts.Force,
			RmiForce:                      opts.Force,
			RmForce:                       true,
			DryRun:                        opts.DryRun,
		}

		if err := logboek.Context(ctx).Default().LogProcess("Running cleanup for docker containers created by werf").DoError(func() error {
			return safeContainersCleanup(ctx, commonOptions)
		}); err != nil {
			return err
		}

		if err := logboek.Context(ctx).Default().LogProcess("Running cleanup for dangling docker images created by werf").DoError(func() error {
			return safeDanglingImagesCleanup(ctx, commonOptions)
		}); err != nil {
			return err
		}

		if freedImagesCount == 0 {
			break
		}
		if opts.DryRun {
			break
		}

		logboek.Context(ctx).Default().LogOptionalLn()
	}

	return nil
}

type LocalImageDesc struct {
	ImageSummary types.ImageSummary
	LastUsedAt   time.Time
}

type ImagesLruSort []*LocalImageDesc

func (a ImagesLruSort) Len() int { return len(a) }
func (a ImagesLruSort) Less(i, j int) bool {
	return a[i].LastUsedAt.Before(a[j].LastUsedAt)
}
func (a ImagesLruSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

type VolumeUsage struct {
	UsedBytes  uint64
	TotalBytes uint64
	Percentage float64
}

func GetVolumeUsageByPath(ctx context.Context, path string) (VolumeUsage, error) {
	di, err := disk.GetInfo(path)
	if err != nil {
		return VolumeUsage{}, fmt.Errorf("unable to get disk info: %s", err)
	}

	usedBytes := di.Total - di.Free
	return VolumeUsage{
		UsedBytes:  usedBytes,
		TotalBytes: di.Total,
		Percentage: (float64(usedBytes) / float64(di.Total)) * 100,
	}, nil
}

func safeDanglingImagesCleanup(ctx context.Context, options CommonOptions) error {
	images, err := werfImagesByFilterSet(ctx, danglingFilterSet())
	if err != nil {
		return err
	}

	var imagesToRemove []types.ImageSummary

	for _, img := range images {
		imagesToRemove = append(imagesToRemove, img)
	}

	imagesToRemove, err = processUsedImages(ctx, imagesToRemove, options)
	if err != nil {
		return err
	}

	if err := imagesRemove(ctx, imagesToRemove, options); err != nil {
		return err
	}

	return nil
}

func safeContainersCleanup(ctx context.Context, options CommonOptions) error {
	containers, err := werfContainersByFilterSet(ctx, filters.NewArgs())
	if err != nil {
		return fmt.Errorf("cannot get stages build containers: %s", err)
	}

	for _, container := range containers {
		var containerName string
		for _, name := range container.Names {
			if strings.HasPrefix(name, fmt.Sprintf("/%s", image.StageContainerNamePrefix)) {
				containerName = strings.TrimPrefix(name, "/")
				break
			}
		}

		if containerName == "" {
			logboek.Context(ctx).Warn().LogF("Ignore bad container %s\n", container.ID)
			continue
		}

		if err := containersRemove(ctx, []types.Container{container}, options); err != nil {
			return fmt.Errorf("failed to remove container %s: %s", logContainerName(container), err)
		}
	}

	return nil
}
