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

	"github.com/werf/kubedog/pkg/utils"
	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/storage/lrumeta"
	"github.com/werf/werf/pkg/volumeutils"
	"github.com/werf/werf/pkg/werf"
)

const (
	MinImagesToDelete = 10
)

func GetLocalDockerServerStoragePath(ctx context.Context) (string, error) {
	dockerInfo, err := docker.Info(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to get docker info: %w", err)
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
		return "", fmt.Errorf("error accessing %q: %w", storagePath, err)
	}
	return storagePath, nil
}

func getDockerServerStoragePath(ctx context.Context, dockerServerStoragePathOption *string) (string, error) {
	var dockerServerStoragePath string
	if dockerServerStoragePathOption != nil && *dockerServerStoragePathOption != "" {
		dockerServerStoragePath = *dockerServerStoragePathOption
	} else {
		path, err := GetLocalDockerServerStoragePath(ctx)
		if err != nil {
			return "", err
		}
		dockerServerStoragePath = path
	}

	return dockerServerStoragePath, nil
}

func ShouldRunAutoGCForLocalDockerServer(ctx context.Context, allowedVolumeUsagePercentage float64, dockerServerStoragePath string) (bool, error) {
	if dockerServerStoragePath == "" {
		return false, nil
	}

	vu, err := volumeutils.GetVolumeUsageByPath(ctx, dockerServerStoragePath)
	if err != nil {
		return false, fmt.Errorf("error getting volume usage by path %q: %w", dockerServerStoragePath, err)
	}

	return vu.Percentage > allowedVolumeUsagePercentage, nil
}

type LocalDockerServerStorageCheckResult struct {
	VolumeUsage      volumeutils.VolumeUsage
	TotalImagesBytes uint64
	ImagesDescs      []*LocalImageDesc
}

func (checkResult *LocalDockerServerStorageCheckResult) GetBytesToFree(targetVolumeUsage float64) uint64 {
	allowedVolumeUsageToFree := checkResult.VolumeUsage.Percentage - targetVolumeUsage
	bytesToFree := uint64((float64(checkResult.VolumeUsage.TotalBytes) / 100.0) * allowedVolumeUsageToFree)
	return bytesToFree
}

func GetLocalDockerServerStorageCheck(ctx context.Context, dockerServerStoragePath string) (*LocalDockerServerStorageCheckResult, error) {
	res := &LocalDockerServerStorageCheckResult{}

	vu, err := volumeutils.GetVolumeUsageByPath(ctx, dockerServerStoragePath)
	if err != nil {
		return nil, fmt.Errorf("error getting volume usage by path %q: %w", dockerServerStoragePath, err)
	}
	res.VolumeUsage = vu

	var images []types.ImageSummary

	{
		filterSet := filters.NewArgs()
		filterSet.Add("label", image.WerfLabel)
		filterSet.Add("label", image.WerfStageDigestLabel)

		imgs, err := docker.Images(ctx, types.ImageListOptions{Filters: filterSet})
		if err != nil {
			return nil, fmt.Errorf("unable to get werf docker images: %w", err)
		}

		// Do not remove stages-storage=:local images, because this is primary stages storage data,
		// and it can only be cleaned by the werf-cleanup command
	ExcludeLocalV1_2StagesStorage:
		for _, img := range imgs {
			projectName := img.Labels[image.WerfLabel]

			for _, ref := range img.RepoTags {
				if strings.HasPrefix(ref, fmt.Sprintf("%s:", projectName)) {
					continue ExcludeLocalV1_2StagesStorage
				}
			}

			images = append(images, img)
		}
	}

	// Process legacy v1.1 images
	{
		filterSet := filters.NewArgs()
		filterSet.Add("label", image.WerfLabel)
		filterSet.Add("label", "werf-stage-signature") // v1.1 legacy images

		imgs, err := docker.Images(ctx, types.ImageListOptions{Filters: filterSet})
		if err != nil {
			return nil, fmt.Errorf("unable to get werf v1.1 legacy docker images: %w", err)
		}

		// Do not remove stages-storage=:local images, because this is primary stages storage data,
		// and it can only be cleaned by the werf-cleanup command
	ExcludeLocalV1_1StagesStorage:
		for _, img := range imgs {
			for _, ref := range img.RepoTags {
				if strings.HasPrefix(ref, "werf-stages-storage/") {
					continue ExcludeLocalV1_1StagesStorage
				}
			}

			images = append(images, img)
		}
	}

	{
		// **NOTICE** Remove v1.1 last-run-at timestamp check when v1.1 reaches its end of life

		t, err := werf.GetWerfLastRunAtV1_1(ctx)
		if err != nil {
			return nil, fmt.Errorf("error getting v1.1 last run timestamp: %w", err)
		}

		// No werf v1.1 runs on this host.
		// This is stupid check, but the only available safe option at the moment.
		if t.IsZero() {
			filterSet := filters.NewArgs()

			filterSet.Add("reference", "*client-id-*")
			filterSet.Add("reference", "*managed-image-*")
			filterSet.Add("reference", "*meta-*")
			filterSet.Add("reference", "*import-metadata-*")
			filterSet.Add("reference", "*-rejected")

			filterSet.Add("reference", "werf-client-id/*")
			filterSet.Add("reference", "werf-managed-images/*")
			filterSet.Add("reference", "werf-images-metadata-by-commit/*")
			filterSet.Add("reference", "werf-import-metadata/*")

			imgs, err := docker.Images(ctx, types.ImageListOptions{Filters: filterSet})
			if err != nil {
				return nil, fmt.Errorf("unable to get werf service images: %w", err)
			}

			for _, img := range imgs {
				// **NOTICE.** Cannot remove by werf label, because currently there is no such label for service-images by historical reasons.
				// So check by size at least for now.
				if img.Size != 0 {
					continue
				}

				images = append(images, img)
			}
		}
	}

CreateImagesDescs:
	for _, imageSummary := range images {
		data, _ := json.Marshal(imageSummary)
		logboek.Context(ctx).Debug().LogF("Image summary:\n%s\n---\n", data)

		res.TotalImagesBytes += uint64(imageSummary.VirtualSize - imageSummary.SharedSize)

		lastUsedAt := time.Unix(imageSummary.Created, 0)

	CheckEachRef:
		for _, ref := range imageSummary.RepoTags {
			// IMPORTANT: ignore none images, these may be either orphans or just built fresh images and we shall not delete these
			if ref == "<none>:<none>" {
				continue CreateImagesDescs
			}

			lastRecentlyUsedAt, err := lrumeta.CommonLRUImagesCache.GetImageLastAccessTime(ctx, ref)
			if err != nil {
				return nil, fmt.Errorf("error accessing last recently used images cache: %w", err)
			}

			if lastRecentlyUsedAt.IsZero() {
				continue CheckEachRef
			}

			lastUsedAt = lastRecentlyUsedAt
			break
		}

		desc := &LocalImageDesc{
			ImageSummary: imageSummary,
			LastUsedAt:   lastUsedAt,
		}
		res.ImagesDescs = append(res.ImagesDescs, desc)
	}

	sort.Sort(ImagesLruSort(res.ImagesDescs))

	return res, nil
}

func RunGCForLocalDockerServer(ctx context.Context, allowedVolumeUsagePercentage, allowedVolumeUsageMarginPercentage float64, dockerServerStoragePath string, force, dryRun bool) error {
	if dockerServerStoragePath == "" {
		return nil
	}

	targetVolumeUsage := allowedVolumeUsagePercentage - allowedVolumeUsageMarginPercentage
	if targetVolumeUsage < 0 {
		targetVolumeUsage = 0
	}

	checkResult, err := GetLocalDockerServerStorageCheck(ctx, dockerServerStoragePath)
	if err != nil {
		return fmt.Errorf("error getting local docker server storage check: %w", err)
	}

	bytesToFree := checkResult.GetBytesToFree(targetVolumeUsage)

	if checkResult.VolumeUsage.Percentage <= allowedVolumeUsagePercentage {
		logboek.Context(ctx).Default().LogBlock("Local docker server storage check").Do(func() {
			logboek.Context(ctx).Default().LogF("Docker server storage path: %s\n", dockerServerStoragePath)
			logboek.Context(ctx).Default().LogF("Volume usage: %s / %s\n", humanize.Bytes(checkResult.VolumeUsage.UsedBytes), humanize.Bytes(checkResult.VolumeUsage.TotalBytes))
			logboek.Context(ctx).Default().LogF("Allowed volume usage percentage: %s <= %s — %s\n", utils.GreenF("%0.2f%%", checkResult.VolumeUsage.Percentage), utils.BlueF("%0.2f%%", allowedVolumeUsagePercentage), utils.GreenF("OK"))
		})

		return nil
	}

	logboek.Context(ctx).Default().LogBlock("Local docker server storage check").Do(func() {
		logboek.Context(ctx).Default().LogF("Docker server storage path: %s\n", dockerServerStoragePath)
		logboek.Context(ctx).Default().LogF("Volume usage: %s / %s\n", humanize.Bytes(checkResult.VolumeUsage.UsedBytes), humanize.Bytes(checkResult.VolumeUsage.TotalBytes))
		logboek.Context(ctx).Default().LogF("Allowed percentage level exceeded: %s > %s — %s\n", utils.RedF("%0.2f%%", checkResult.VolumeUsage.Percentage), utils.YellowF("%0.2f%%", allowedVolumeUsagePercentage), utils.RedF("HIGH VOLUME USAGE"))
		logboek.Context(ctx).Default().LogF("Target percentage level after cleanup: %0.2f%% - %0.2f%% (margin) = %s\n", allowedVolumeUsagePercentage, allowedVolumeUsageMarginPercentage, utils.BlueF("%0.2f%%", targetVolumeUsage))
		logboek.Context(ctx).Default().LogF("Needed to free: %s\n", utils.RedF("%s", humanize.Bytes(bytesToFree)))
		logboek.Context(ctx).Default().LogF("Available images to free: %s\n", utils.YellowF("%d", len(checkResult.ImagesDescs)))
	})

	var processedDockerImagesIDs []string
	var processedDockerContainersIDs []string

	for {
		var freedBytes uint64
		var freedImagesCount uint64
		var acquiredHostLocks []lockgate.LockHandle

		if len(checkResult.ImagesDescs) > 0 {
			if err := logboek.Context(ctx).Default().LogProcess("Running cleanup for least recently used docker images created by werf").DoError(func() error {
			DeleteImages:
				for _, desc := range checkResult.ImagesDescs {
					for _, id := range processedDockerImagesIDs {
						if desc.ImageSummary.ID == id {
							logboek.Context(ctx).Default().LogFDetails("Skip already processed image %q\n", desc.ImageSummary.ID)
							continue DeleteImages
						}
					}
					processedDockerImagesIDs = append(processedDockerImagesIDs, desc.ImageSummary.ID)

					imageRemoved := false

					if len(desc.ImageSummary.RepoTags) > 0 {
						allTagsRemoved := true

						for _, ref := range desc.ImageSummary.RepoTags {
							if ref == "<none>:<none>" {
								if err := removeImage(ctx, desc.ImageSummary.ID, force, dryRun); err != nil {
									logboek.Context(ctx).Warn().LogF("failed to remove local docker image by ID %q: %s\n", desc.ImageSummary.ID, err)
									allTagsRemoved = false
								}
							} else {
								lockName := container_backend.ImageLockName(ref)

								isLocked, lock, err := werf.AcquireHostLock(ctx, lockName, lockgate.AcquireOptions{NonBlocking: true})
								if err != nil {
									return fmt.Errorf("error locking image %q: %w", lockName, err)
								}

								if !isLocked {
									logboek.Context(ctx).Default().LogFDetails("Image %q is locked at the moment: skip removal\n", ref)
									continue DeleteImages
								}

								acquiredHostLocks = append(acquiredHostLocks, lock)

								if err := removeImage(ctx, ref, force, dryRun); err != nil {
									logboek.Context(ctx).Warn().LogF("failed to remove local docker image by repo tag %q: %s\n", ref, err)
									allTagsRemoved = false
								}
							}
						}

						if allTagsRemoved {
							imageRemoved = true
						}
					} else if len(desc.ImageSummary.RepoDigests) > 0 {
						allDigestsRemoved := true

						for _, repoDigest := range desc.ImageSummary.RepoDigests {
							if err := removeImage(ctx, repoDigest, force, dryRun); err != nil {
								logboek.Context(ctx).Warn().LogF("failed to remove local docker image by repo digest %q: %s\n", repoDigest, err)
								allDigestsRemoved = false
							}
						}

						if allDigestsRemoved {
							imageRemoved = true
						}
					}

					if imageRemoved {
						freedBytes += uint64(desc.ImageSummary.VirtualSize - desc.ImageSummary.SharedSize)
						freedImagesCount++
					}

					if freedImagesCount < MinImagesToDelete {
						continue
					}

					if freedBytes > bytesToFree {
						break
					}
				}

				logboek.Context(ctx).Default().LogF("Freed images: %s\n", utils.GreenF("%d", freedImagesCount))

				return nil
			}); err != nil {
				return err
			}
		}

		if freedImagesCount == 0 {
			logboek.Context(ctx).Warn().LogF("WARNING: Detected high docker storage volume usage, while no werf images available to cleanup!\n")
			logboek.Context(ctx).Warn().LogF("WARNING:\n")
			logboek.Context(ctx).Warn().LogF("WARNING: Werf tries to maintain host clean by deleting:\n")
			logboek.Context(ctx).Warn().LogF("WARNING:  - old unused files from werf caches (which are stored in the ~/.werf/local_cache);\n")
			logboek.Context(ctx).Warn().LogF("WARNING:  - old temporary service files /tmp/werf-project-data-* and /tmp/werf-config-render-*;\n")
			logboek.Context(ctx).Warn().LogF("WARNING:  - least recently used werf images except local stages storage images (images built with 'werf build' without '--repo' param, or with '--stages-storage=:local' param for the werf v1.1).\n")
			logboek.Context(ctx).Warn().LogOptionalLn()
		}

		for _, lock := range acquiredHostLocks {
			if err := werf.ReleaseHostLock(lock); err != nil {
				return fmt.Errorf("unable to release lock %q: %w", lock.LockName, err)
			}
		}

		commonOptions := CommonOptions{
			RmContainersThatUseWerfImages: force,
			SkipUsedImages:                !force,
			RmiForce:                      force,
			RmForce:                       true,
			DryRun:                        dryRun,
		}

		if err := logboek.Context(ctx).Default().LogProcess("Running cleanup for docker containers created by werf").DoError(func() error {
			newProcessedContainersIDs, err := safeContainersCleanup(ctx, processedDockerContainersIDs, commonOptions)
			if err != nil {
				return fmt.Errorf("safe containers cleanup failed: %w", err)
			}

			processedDockerContainersIDs = newProcessedContainersIDs

			return nil
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
		if dryRun {
			break
		}

		logboek.Context(ctx).Default().LogOptionalLn()

		checkResult, err = GetLocalDockerServerStorageCheck(ctx, dockerServerStoragePath)
		if err != nil {
			return fmt.Errorf("error getting local docker server storage check: %w", err)
		}

		if checkResult.VolumeUsage.Percentage <= targetVolumeUsage {
			logboek.Context(ctx).Default().LogBlock("Local docker server storage check").Do(func() {
				logboek.Context(ctx).Default().LogF("Docker server storage path: %s\n", dockerServerStoragePath)
				logboek.Context(ctx).Default().LogF("Volume usage: %s / %s\n", humanize.Bytes(checkResult.VolumeUsage.UsedBytes), humanize.Bytes(checkResult.VolumeUsage.TotalBytes))
				logboek.Context(ctx).Default().LogF("Target volume usage percentage: %s <= %s — %s\n", utils.GreenF("%0.2f%%", checkResult.VolumeUsage.Percentage), utils.BlueF("%0.2f%%", targetVolumeUsage), utils.GreenF("OK"))
			})

			break
		}

		bytesToFree = checkResult.GetBytesToFree(targetVolumeUsage)

		logboek.Context(ctx).Default().LogBlock("Local docker server storage check").Do(func() {
			logboek.Context(ctx).Default().LogF("Docker server storage path: %s\n", dockerServerStoragePath)
			logboek.Context(ctx).Default().LogF("Volume usage: %s / %s\n", humanize.Bytes(checkResult.VolumeUsage.UsedBytes), humanize.Bytes(checkResult.VolumeUsage.TotalBytes))
			logboek.Context(ctx).Default().LogF("Target volume usage percentage: %s > %s — %s\n", utils.RedF("%0.2f%%", checkResult.VolumeUsage.Percentage), utils.BlueF("%0.2f%%", targetVolumeUsage), utils.RedF("HIGH VOLUME USAGE"))
			logboek.Context(ctx).Default().LogF("Needed to free: %s\n", utils.RedF("%s", humanize.Bytes(bytesToFree)))
			logboek.Context(ctx).Default().LogF("Available images to free: %s\n", utils.YellowF("%d", len(checkResult.ImagesDescs)))
		})
	}

	return nil
}

func removeImage(ctx context.Context, ref string, force, dryRun bool) error {
	logboek.Context(ctx).Default().LogF("Removing %s\n", ref)
	if dryRun {
		return nil
	}

	args := []string{ref}
	if force {
		args = append(args, "--force")
	}

	return docker.CliRmi(ctx, args...)
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

func safeDanglingImagesCleanup(ctx context.Context, options CommonOptions) error {
	images, err := trueDanglingImages(ctx)
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

func safeContainersCleanup(ctx context.Context, processedDockerContainersIDs []string, options CommonOptions) ([]string, error) {
	containers, err := werfContainersByFilterSet(ctx, filters.NewArgs())
	if err != nil {
		return nil, fmt.Errorf("cannot get stages build containers: %w", err)
	}

ProcessContainers:
	for _, container := range containers {
		for _, id := range processedDockerContainersIDs {
			if id == container.ID {
				continue ProcessContainers
			}
		}
		processedDockerContainersIDs = append(processedDockerContainersIDs, container.ID)

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

		if err := func() error {
			containerLockName := container_backend.ContainerLockName(containerName)
			isLocked, lock, err := werf.AcquireHostLock(ctx, containerLockName, lockgate.AcquireOptions{NonBlocking: true})
			if err != nil {
				return fmt.Errorf("failed to lock %s for container %s: %w", containerLockName, logContainerName(container), err)
			}

			if !isLocked {
				logboek.Context(ctx).Default().LogFDetails("Ignore container %s used by another process\n", logContainerName(container))
				return nil
			}
			defer werf.ReleaseHostLock(lock)

			if err := containersRemove(ctx, []types.Container{container}, options); err != nil {
				return fmt.Errorf("failed to remove container %s: %w", logContainerName(container), err)
			}

			return nil
		}(); err != nil {
			return nil, err
		}
	}

	return processedDockerContainersIDs, nil
}
