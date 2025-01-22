package host_cleaning

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/werf/common-go/pkg/lock"
	"github.com/werf/common-go/pkg/util"
	"github.com/werf/kubedog/pkg/utils"
	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage/lrumeta"
	"github.com/werf/werf/v2/pkg/volumeutils"
	"github.com/werf/werf/v2/pkg/werf"
)

var ErrUnsupportedContainerBackend = errors.New("unsupported container backend")

type RunGCOptions struct {
	AllowedStorageVolumeUsagePercentage       float64
	AllowedStorageVolumeUsageMarginPercentage float64
	StoragePath                               string
	Force                                     bool
	DryRun                                    bool
}

type RunAutoGCOptions struct {
	AllowedStorageVolumeUsagePercentage float64
	StoragePath                         string
}

type LocalBackendCleaner struct {
	backend           container_backend.ContainerBackend
	backendName       string
	minImagesToDelete uint64
	// refs for stubbing in testing
	volumeutilsGetVolumeUsageByPath func(ctx context.Context, path string) (volumeutils.VolumeUsage, error)
	werfGetWerfLastRunAtV1_1        func(ctx context.Context) (time.Time, error)
	lrumetaGetImageLastAccessTime   func(ctx context.Context, imageRef string) (time.Time, error)
}

func NewLocalBackendCleaner(backend container_backend.ContainerBackend) (*LocalBackendCleaner, error) {
	cleaner := &LocalBackendCleaner{
		backend:           backend,
		minImagesToDelete: 10,
		// refs for stubbing in testing
		volumeutilsGetVolumeUsageByPath: volumeutils.GetVolumeUsageByPath,
		werfGetWerfLastRunAtV1_1:        werf.GetWerfLastRunAtV1_1,
		lrumetaGetImageLastAccessTime:   lrumeta.CommonLRUImagesCache.GetImageLastAccessTime,
	}

	switch backend.(type) {
	case *container_backend.DockerServerBackend:
		cleaner.backendName = "docker"
		return cleaner, nil
	case *container_backend.BuildahBackend:
		cleaner.backendName = "buildah"
		return cleaner, nil
	default:
		// returns cleaner for testing with mock
		cleaner.backendName = "test"
		return cleaner, ErrUnsupportedContainerBackend
	}
}

func (cleaner *LocalBackendCleaner) BackendName() string {
	return cleaner.backendName
}

func (cleaner *LocalBackendCleaner) backendStoragePath(ctx context.Context, storagePath string) (string, error) {
	backendStoragePath := storagePath

	if backendStoragePath == "" {
		info, err := cleaner.backend.Info(ctx)
		if err != nil {
			return "", fmt.Errorf("errot getting local %s backend info: %w", cleaner.BackendName(), err)
		}
		backendStoragePath = info.StoreGraphRoot
	}

	// assert permissions
	if _, err := os.Stat(backendStoragePath); os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("error accessing %q: %w", backendStoragePath, err)
	}

	return backendStoragePath, nil
}

func (cleaner *LocalBackendCleaner) ShouldRunAutoGC(ctx context.Context, options RunAutoGCOptions) (bool, error) {
	backendStoragePath, err := cleaner.backendStoragePath(ctx, options.StoragePath)
	if err != nil {
		return false, fmt.Errorf("error getting local %s backend storage path: %w", cleaner.BackendName(), err)
	}

	vu, err := cleaner.volumeutilsGetVolumeUsageByPath(ctx, backendStoragePath)
	if err != nil {
		return false, fmt.Errorf("error getting volume usage by path %q: %w", backendStoragePath, err)
	}

	return vu.Percentage > options.AllowedStorageVolumeUsagePercentage, nil
}

type CheckResultBackendStorage struct {
	VolumeUsage      volumeutils.VolumeUsage
	TotalImagesBytes uint64
	ImagesDescs      []*LocalImageDesc
}

func (checkResult *CheckResultBackendStorage) GetBytesToFree(targetVolumeUsage float64) uint64 {
	allowedVolumeUsageToFree := checkResult.VolumeUsage.Percentage - targetVolumeUsage
	bytesToFree := uint64((float64(checkResult.VolumeUsage.TotalBytes) / 100.0) * allowedVolumeUsageToFree)
	return bytesToFree
}

func (cleaner *LocalBackendCleaner) checkBackendStorage(ctx context.Context, backendStoragePath string) (*CheckResultBackendStorage, error) {
	res := &CheckResultBackendStorage{}

	vu, err := cleaner.volumeutilsGetVolumeUsageByPath(ctx, backendStoragePath)
	if err != nil {
		return nil, fmt.Errorf("error getting volume usage by path %q: %w", backendStoragePath, err)
	}
	res.VolumeUsage = vu

	var images image.ImagesList

	{

		imgs, err := cleaner.backend.Images(ctx, buildImagesOptions(
			util.NewPair("label", image.WerfLabel),
			util.NewPair("label", image.WerfStageDigestLabel),
		))
		if err != nil {
			return nil, fmt.Errorf("unable to get werf %s images: %w", cleaner.BackendName(), err)
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
		imgs, err := cleaner.backend.Images(ctx, buildImagesOptions(
			util.NewPair("label", image.WerfLabel),
			util.NewPair("label", "werf-stage-signature"), // v1.1 legacy images
		))
		if err != nil {
			return nil, fmt.Errorf("unable to get werf v1.1 legacy %s images: %w", cleaner.BackendName(), err)
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

		t, err := cleaner.werfGetWerfLastRunAtV1_1(ctx)
		if err != nil {
			return nil, fmt.Errorf("error getting v1.1 last run timestamp: %w", err)
		}

		// No werf v1.1 runs on this host.
		// This is stupid check, but the only available safe option at the moment.
		if t.IsZero() {
			imgs, err := cleaner.backend.Images(ctx, buildImagesOptions(
				util.NewPair("reference", "*client-id-*"),
				util.NewPair("reference", "*managed-image-*"),
				util.NewPair("reference", "*meta-*"),
				util.NewPair("reference", "*import-metadata-*"),
				util.NewPair("reference", "*-rejected"),

				util.NewPair("reference", "werf-client-id/*"),
				util.NewPair("reference", "werf-managed-images/*"),
				util.NewPair("reference", "werf-images-metadata-by-commit/*"),
				util.NewPair("reference", "werf-import-metadata/*"),
			))
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

		res.TotalImagesBytes += uint64(imageSummary.Size - normalizeSharedSize(imageSummary.SharedSize))

		lastUsedAt := imageSummary.Created

	CheckEachRef:
		for _, ref := range imageSummary.RepoTags {
			// IMPORTANT: ignore none images, these may be either orphans or just built fresh images and we shall not delete these
			if ref == "<none>:<none>" {
				continue CreateImagesDescs
			}

			lastRecentlyUsedAt, err := cleaner.lrumetaGetImageLastAccessTime(ctx, ref)
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

	slices.SortFunc(res.ImagesDescs, func(a, b *LocalImageDesc) int {
		return a.LastUsedAt.Compare(b.LastUsedAt)
	})

	return res, nil
}

func (cleaner *LocalBackendCleaner) RunGC(ctx context.Context, options RunGCOptions) error {
	backendStoragePath, err := cleaner.backendStoragePath(ctx, options.StoragePath)
	if err != nil {
		return fmt.Errorf("error getting local %s backend storage path: %w", cleaner.BackendName(), err)
	}

	targetVolumeUsage := math.Max(options.AllowedStorageVolumeUsagePercentage-options.AllowedStorageVolumeUsageMarginPercentage, 0)

	checkResult, err := cleaner.checkBackendStorage(ctx, backendStoragePath)
	if err != nil {
		return fmt.Errorf("error getting local %s backend storage check: %w", cleaner.BackendName(), err)
	}

	if checkResult.VolumeUsage.Percentage <= options.AllowedStorageVolumeUsagePercentage {
		logboek.Context(ctx).Default().LogBlock("Local %s backend storage check", cleaner.BackendName()).Do(func() {
			logboek.Context(ctx).Default().LogF("Storage path: %s\n", backendStoragePath)
			logboek.Context(ctx).Default().LogF("Volume usage: %s / %s\n", humanize.Bytes(checkResult.VolumeUsage.UsedBytes), humanize.Bytes(checkResult.VolumeUsage.TotalBytes))
			logboek.Context(ctx).Default().LogF("Allowed volume usage percentage: %s <= %s — %s\n", utils.GreenF("%0.2f%%", checkResult.VolumeUsage.Percentage), utils.BlueF("%0.2f%%", options.AllowedStorageVolumeUsagePercentage), utils.GreenF("OK"))
		})

		return nil
	}

	bytesToFree := checkResult.GetBytesToFree(targetVolumeUsage)

	logboek.Context(ctx).Default().LogBlock("Local %s backend storage check", cleaner.BackendName()).Do(func() {
		logboek.Context(ctx).Default().LogF("Storage path: %s\n", backendStoragePath)
		logboek.Context(ctx).Default().LogF("Volume usage: %s / %s\n", humanize.Bytes(checkResult.VolumeUsage.UsedBytes), humanize.Bytes(checkResult.VolumeUsage.TotalBytes))
		logboek.Context(ctx).Default().LogF("Allowed percentage level exceeded: %s > %s — %s\n", utils.RedF("%0.2f%%", checkResult.VolumeUsage.Percentage), utils.YellowF("%0.2f%%", options.AllowedStorageVolumeUsagePercentage), utils.RedF("HIGH VOLUME USAGE"))
		logboek.Context(ctx).Default().LogF("Target percentage level after cleanup: %0.2f%% - %0.2f%% (margin) = %s\n", options.AllowedStorageVolumeUsagePercentage, options.AllowedStorageVolumeUsageMarginPercentage, utils.BlueF("%0.2f%%", targetVolumeUsage))
		logboek.Context(ctx).Default().LogF("Needed to free: %s\n", utils.RedF("%s", humanize.Bytes(bytesToFree)))
		logboek.Context(ctx).Default().LogF("Available images to free: %s\n", utils.YellowF("%d", len(checkResult.ImagesDescs)))
	})

	var processedImagesIDs []string
	var processedContainersIDs []string

	for {
		var freedBytes uint64
		var freedImagesCount uint64
		var acquiredHostLocks []lockgate.LockHandle

		if len(checkResult.ImagesDescs) > 0 {
			if err := logboek.Context(ctx).Default().LogProcess("Running cleanup for least recently used %s images created by werf", cleaner.BackendName()).DoError(func() error {
			DeleteImages:
				for _, desc := range checkResult.ImagesDescs {
					for _, id := range processedImagesIDs {
						if desc.ImageSummary.ID == id {
							logboek.Context(ctx).Default().LogFDetails("Skip already processed image %q\n", desc.ImageSummary.ID)
							continue DeleteImages
						}
					}
					processedImagesIDs = append(processedImagesIDs, desc.ImageSummary.ID)

					imageRemoved := false

					if len(desc.ImageSummary.RepoTags) > 0 {
						allTagsRemoved := true

						for _, ref := range desc.ImageSummary.RepoTags {
							if ref == "<none>:<none>" {
								if err := cleaner.removeImage(ctx, desc.ImageSummary.ID, options.Force, options.DryRun); err != nil {
									logboek.Context(ctx).Warn().LogF("failed to remove local %s image by ID %q: %s\n", cleaner.BackendName(), desc.ImageSummary.ID, err)
									allTagsRemoved = false
								}
							} else {
								lockName := container_backend.ImageLockName(ref)

								isLocked, lock, err := chart.AcquireHostLock(ctx, lockName, lockgate.AcquireOptions{NonBlocking: true})
								if err != nil {
									return fmt.Errorf("error locking image %q: %w", lockName, err)
								}

								if !isLocked {
									logboek.Context(ctx).Default().LogFDetails("Image %q is locked at the moment: skip removal\n", ref)
									continue DeleteImages
								}

								acquiredHostLocks = append(acquiredHostLocks, lock)

								if err := cleaner.removeImage(ctx, ref, options.Force, options.DryRun); err != nil {
									logboek.Context(ctx).Warn().LogF("failed to remove local %s image by repo tag %q: %s\n", cleaner.BackendName(), ref, err)
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
							if err := cleaner.removeImage(ctx, repoDigest, options.Force, options.DryRun); err != nil {
								logboek.Context(ctx).Warn().LogF("failed to remove local %s image by repo digest %q: %s\n", cleaner.BackendName(), repoDigest, err)
								allDigestsRemoved = false
							}
						}

						if allDigestsRemoved {
							imageRemoved = true
						}
					}

					if imageRemoved {
						freedBytes += uint64(desc.ImageSummary.Size - normalizeSharedSize(desc.ImageSummary.SharedSize))
						freedImagesCount++
					}

					if freedImagesCount < cleaner.minImagesToDelete {
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
			logboek.Context(ctx).Warn().LogF("WARNING: Detected high %s storage volume usage, while no werf images available to cleanup!\n", cleaner.BackendName())
			logboek.Context(ctx).Warn().LogF("WARNING:\n")
			logboek.Context(ctx).Warn().LogF("WARNING: Werf tries to maintain host clean by deleting:\n")
			logboek.Context(ctx).Warn().LogF("WARNING:  - old unused files from werf caches (which are stored in the ~/.werf/local_cache);\n")
			logboek.Context(ctx).Warn().LogF("WARNING:  - old temporary service files /tmp/werf-project-data-* and /tmp/werf-config-render-*;\n")
			logboek.Context(ctx).Warn().LogF("WARNING:  - least recently used werf images except local stages storage images (images built with 'werf build' without '--repo' param, or with '--stages-storage=:local' param for the werf v1.1).\n")
			logboek.Context(ctx).Warn().LogOptionalLn()
		}

		for _, lock := range acquiredHostLocks {
			if err := chart.ReleaseHostLock(lock); err != nil {
				return fmt.Errorf("unable to release lock %q: %w", lock.LockName, err)
			}
		}

		commonOptions := CommonOptions{
			RmContainersThatUseWerfImages: options.Force,
			SkipUsedImages:                !options.Force,
			RmiForce:                      options.Force,
			RmForce:                       true,
			DryRun:                        options.DryRun,
		}

		if err := logboek.Context(ctx).Default().LogProcess("Running cleanup for %s containers created by werf", cleaner.BackendName()).DoError(func() error {
			newProcessedContainersIDs, err := cleaner.safeContainersCleanup(ctx, processedContainersIDs, commonOptions)
			if err != nil {
				return fmt.Errorf("safe containers cleanup failed: %w", err)
			}

			processedContainersIDs = newProcessedContainersIDs

			return nil
		}); err != nil {
			return err
		}

		if err := logboek.Context(ctx).Default().LogProcess("Running cleanup for dangling %s images created by werf", cleaner.BackendName()).DoError(func() error {
			return cleaner.safeDanglingImagesCleanup(ctx, commonOptions)
		}); err != nil {
			return err
		}

		if freedImagesCount == 0 {
			break
		}
		if options.DryRun {
			break
		}

		logboek.Context(ctx).Default().LogOptionalLn()

		checkResult, err = cleaner.checkBackendStorage(ctx, backendStoragePath)
		if err != nil {
			return fmt.Errorf("error getting local %s backend storage check: %w", cleaner.BackendName(), err)
		}

		if checkResult.VolumeUsage.Percentage <= targetVolumeUsage {
			logboek.Context(ctx).Default().LogBlock("Local %s backend storage check", cleaner.BackendName()).Do(func() {
				logboek.Context(ctx).Default().LogF("Storage path: %s\n", backendStoragePath)
				logboek.Context(ctx).Default().LogF("Volume usage: %s / %s\n", humanize.Bytes(checkResult.VolumeUsage.UsedBytes), humanize.Bytes(checkResult.VolumeUsage.TotalBytes))
				logboek.Context(ctx).Default().LogF("Target volume usage percentage: %s <= %s — %s\n", utils.GreenF("%0.2f%%", checkResult.VolumeUsage.Percentage), utils.BlueF("%0.2f%%", targetVolumeUsage), utils.GreenF("OK"))
			})

			break
		}

		bytesToFree = checkResult.GetBytesToFree(targetVolumeUsage)

		logboek.Context(ctx).Default().LogBlock("Local %s backend storage check", cleaner.BackendName()).Do(func() {
			logboek.Context(ctx).Default().LogF("Storage path: %s\n", backendStoragePath)
			logboek.Context(ctx).Default().LogF("Volume usage: %s / %s\n", humanize.Bytes(checkResult.VolumeUsage.UsedBytes), humanize.Bytes(checkResult.VolumeUsage.TotalBytes))
			logboek.Context(ctx).Default().LogF("Target volume usage percentage: %s > %s — %s\n", utils.RedF("%0.2f%%", checkResult.VolumeUsage.Percentage), utils.BlueF("%0.2f%%", targetVolumeUsage), utils.RedF("HIGH VOLUME USAGE"))
			logboek.Context(ctx).Default().LogF("Needed to free: %s\n", utils.RedF("%s", humanize.Bytes(bytesToFree)))
			logboek.Context(ctx).Default().LogF("Available images to free: %s\n", utils.YellowF("%d", len(checkResult.ImagesDescs)))
		})
	}

	return nil
}

func (cleaner *LocalBackendCleaner) removeImage(ctx context.Context, ref string, force, dryRun bool) error {
	logboek.Context(ctx).Default().LogF("Removing %s\n", ref)

	if dryRun {
		return nil
	}

	return cleaner.backend.Rmi(ctx, ref, container_backend.RmiOpts{
		Force: force,
	})
}

type LocalImageDesc struct {
	ImageSummary image.Summary
	LastUsedAt   time.Time
}

func (cleaner *LocalBackendCleaner) safeDanglingImagesCleanup(ctx context.Context, options CommonOptions) error {
	images, err := trueDanglingImages(ctx, cleaner.backend)
	if err != nil {
		return err
	}

	var imagesToRemove image.ImagesList
	for _, img := range images {
		imagesToRemove = append(imagesToRemove, img)
	}

	imagesToRemove, err = processUsedImages(ctx, cleaner.backend, imagesToRemove, options)
	if err != nil {
		return err
	}

	if err := imagesRemove(ctx, cleaner.backend, imagesToRemove, options); err != nil {
		return err
	}

	return nil
}

func (cleaner *LocalBackendCleaner) safeContainersCleanup(ctx context.Context, processedContainersIDs []string, options CommonOptions) ([]string, error) {
	containers, err := werfContainersByContainersOptions(ctx, cleaner.backend, buildContainersOptions())
	if err != nil {
		return nil, fmt.Errorf("cannot get stages build containers: %w", err)
	}

ProcessContainers:
	for _, container := range containers {
		for _, id := range processedContainersIDs {
			if id == container.ID {
				continue ProcessContainers
			}
		}
		processedContainersIDs = append(processedContainersIDs, container.ID)

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
			isLocked, lock, err := chart.AcquireHostLock(ctx, containerLockName, lockgate.AcquireOptions{NonBlocking: true})
			if err != nil {
				return fmt.Errorf("failed to lock %s for container %s: %w", containerLockName, logContainerName(container), err)
			}

			if !isLocked {
				logboek.Context(ctx).Default().LogFDetails("Ignore container %s used by another process\n", logContainerName(container))
				return nil
			}
			defer chart.ReleaseHostLock(lock)

			if err := containersRemove(ctx, cleaner.backend, []image.Container{container}, options); err != nil {
				return fmt.Errorf("failed to remove container %s: %w", logContainerName(container), err)
			}

			return nil
		}(); err != nil {
			return nil, err
		}
	}

	return processedContainersIDs, nil
}

// normalizeSharedSize SharedSize is not calculated by default. `-1` indicates that the value has not been set / calculated.
//
// See https://pkg.go.dev/github.com/docker/docker/api/types/image@v25.0.5+incompatible#Summary.SharedSize
func normalizeSharedSize(size int64) int64 {
	if size == -1 {
		return 0
	}
	return size
}
