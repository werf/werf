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

	"github.com/containers/image/v5/docker/reference"
	"github.com/dustin/go-humanize"

	"github.com/werf/common-go/pkg/lock"
	"github.com/werf/common-go/pkg/util"
	"github.com/werf/kubedog/pkg/utils"
	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/container_backend/prune"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage/lrumeta"
	"github.com/werf/werf/v2/pkg/volumeutils"
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
	backendType       containerBackendType
	minImagesToDelete uint64
	// refs for stubbing in testing
	volumeutilsGetVolumeUsageByPath func(ctx context.Context, path string) (volumeutils.VolumeUsage, error)
	lrumetaGetImageLastAccessTime   func(ctx context.Context, imageRef string) (time.Time, error)
}

func NewLocalBackendCleaner(backend container_backend.ContainerBackend) (*LocalBackendCleaner, error) {
	cleaner := &LocalBackendCleaner{
		backend:           backend,
		minImagesToDelete: 10,
		// refs for stubbing in testing
		volumeutilsGetVolumeUsageByPath: volumeutils.GetVolumeUsageByPath,
		lrumetaGetImageLastAccessTime:   lrumeta.CommonLRUImagesCache.GetImageLastAccessTime,
	}

	switch backend.(type) {
	case *container_backend.DockerServerBackend:
		cleaner.backendType = containerBackendDocker
		return cleaner, nil
	case *container_backend.BuildahBackend:
		cleaner.backendType = containerBackendBuildah
		return cleaner, nil
	default:
		// returns cleaner for testing with mock
		cleaner.backendType = containerBackendTest
		return cleaner, ErrUnsupportedContainerBackend
	}
}

func (cleaner *LocalBackendCleaner) BackendName() string {
	return cleaner.backendType.String()
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
	return vu.Percentage() > options.AllowedStorageVolumeUsagePercentage, nil
}

// safeWerfImages returns werf images are safe for removing
func (cleaner *LocalBackendCleaner) safeWerfImages(ctx context.Context) (image.ImagesList, error) {
	images, err := cleaner.backend.Images(ctx, buildImagesOptions(
		util.NewPair("label", image.WerfLabel),
		util.NewPair("label", image.WerfStageDigestLabel),
	))
	if err != nil {
		return nil, fmt.Errorf("unable to get werf %s images: %w", cleaner.BackendName(), err)
	}

	images, err = cleaner.filterOutImages(ctx, images)
	if err != nil {
		return nil, fmt.Errorf("unable to filter out images: %w", err)
	}

	lastUsedAtMap := make(map[string]time.Time, len(images))

	for _, img := range images {
		data, _ := json.Marshal(img)
		logboek.Context(ctx).Debug().LogF("Image summary:\n%s\n---\n", data)

		lastUsedAtMap[img.ID], err = cleaner.maxLastUsedAtForImage(ctx, img)
		if err != nil {
			return nil, fmt.Errorf("unable to max last used at for image: %w", err)
		}
	}

	slices.SortFunc(images, func(a, b image.Summary) int {
		aTime := lastUsedAtMap[a.ID]
		bTime := lastUsedAtMap[a.ID]
		return aTime.Compare(bTime)
	})

	return images, nil
}

func (cleaner *LocalBackendCleaner) filterOutImages(_ context.Context, images image.ImagesList) (image.ImagesList, error) {
	var list image.ImagesList

skipImage:
	for _, img := range images {
		projectName := img.Labels[image.WerfLabel]

		for _, ref := range img.RepoTags {
			// Do not remove <none>:<none> images.
			// Note: <none>:<none> images are dangling or intermediate images.
			// Right now we don't know what kind of <none>:<none> image is.
			// But we assume backend native garbage collector removes dangling images.
			if ref == "<none>:<none>" {
				continue skipImage
			}

			normalizedTag, err := cleaner.normalizeReference(ref)
			if err != nil {
				return nil, err
			}

			// Do not remove stages-storage=:local images, because this is primary stages storage data,
			// and it can only be cleaned by the werf-cleanup command
			if strings.HasPrefix(normalizedTag, fmt.Sprintf("%s:", projectName)) {
				continue skipImage
			}
		}

		list = append(list, img)
	}

	return list, nil
}

func (cleaner *LocalBackendCleaner) maxLastUsedAtForImage(ctx context.Context, img image.Summary) (time.Time, error) {
	lastUsedAt := img.Created

	for _, ref := range img.RepoTags {
		lastRecentlyUsedAt, err := cleaner.lrumetaGetImageLastAccessTime(ctx, ref)
		if err != nil {
			return time.Time{}, fmt.Errorf("error accessing last recently used images cache: %w", err)
		}

		if lastUsedAt.Before(lastRecentlyUsedAt) {
			lastUsedAt = lastRecentlyUsedAt
		}
	}

	return lastUsedAt, nil
}

// normalizeReference Normalizes image reference (repository tag) to docker backend repository tag format.
func (cleaner *LocalBackendCleaner) normalizeReference(ref string) (string, error) {
	switch cleaner.backendType {
	case containerBackendDocker, containerBackendTest:
		return ref, nil
	case containerBackendBuildah:
		// ------------
		// WORKAROUND for Buildah
		// ------------
		// Buildah repository tag contains hostname (domain) prefix currently.
		// Example (buildah repo tag): localhost/werf-guide-app:e5c6ebcd2718ccfe74d01069a0d758e03d5a2554155ccdc01be0daff-1739265965865
		// We need normalize the tag to docker image repository tag format because our host cleanup algorithm based on docker backend.
		// Example: (docker repo tag): werf-guide-app:e5c6ebcd2718ccfe74d01069a0d758e03d5a2554155ccdc01be0daff-1739265936011
		// https://flant.kaiten.ru/space/193531/boards/card/26364854?focus=comment&focusId=56944076
		// TODO: unify repository tags in v3 on build stage

		named, err := reference.ParseNamed(ref)
		if err != nil {
			return "", err
		}
		hostnamePrefix := fmt.Sprintf("%s/", reference.Domain(named))

		return strings.TrimPrefix(ref, hostnamePrefix), nil
	default:
		return "", ErrUnsupportedContainerBackend
	}
}

func (cleaner *LocalBackendCleaner) RunGC(ctx context.Context, options RunGCOptions) error {
	backendStoragePath, err := cleaner.backendStoragePath(ctx, options.StoragePath)
	if err != nil {
		return fmt.Errorf("error getting local %s backend storage path: %w", cleaner.BackendName(), err)
	}

	logboek.Context(ctx).Default().LogF("Storage path: %s\n", backendStoragePath)
	logboek.Context(ctx).Default().LogOptionalLn()

	vu, err := cleaner.volumeutilsGetVolumeUsageByPath(ctx, backendStoragePath)
	if err != nil {
		return fmt.Errorf("error getting volume usage by path %q: %w", backendStoragePath, err)
	}

	if vu.Percentage() <= options.AllowedStorageVolumeUsagePercentage {
		logboek.Context(ctx).Default().LogBlock("Check storage").Do(func() {
			logboek.Context(ctx).Default().LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
			logboek.Context(ctx).Default().LogF("Allowed volume usage percentage: %s <= %s — %s\n", utils.GreenF("%0.2f%%", vu.Percentage()), utils.BlueF("%0.2f%%", options.AllowedStorageVolumeUsagePercentage), utils.GreenF("OK"))
		})

		return nil
	}

	targetVolumeUsage := uint64(math.Max(options.AllowedStorageVolumeUsagePercentage-options.AllowedStorageVolumeUsageMarginPercentage, 0))

	logboek.Context(ctx).Default().LogBlock("Check storage").Do(func() {
		logboek.Context(ctx).Default().LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
		logboek.Context(ctx).Default().LogF("Allowed percentage level exceeded: %s > %s — %s\n", utils.RedF("%0.2f%%", vu.Percentage()), utils.YellowF("%0.2f%%", options.AllowedStorageVolumeUsagePercentage), utils.RedF("HIGH VOLUME USAGE"))
		logboek.Context(ctx).Default().LogF("Target percentage level after cleanup: %0.2f%% - %0.2f%% (margin) = %s\n", options.AllowedStorageVolumeUsagePercentage, options.AllowedStorageVolumeUsageMarginPercentage, utils.BlueF("%d", targetVolumeUsage))
		logboek.Context(ctx).Default().LogF("Needed to free: %s\n", utils.RedF("%s", humanize.Bytes(calculateBytesToFree(vu, targetVolumeUsage))))
	})

	// Step 1. Prune unused build cache
	reportBuildCache, err := cleaner.pruneBuildCache(ctx, options, vu.UsedBytes, targetVolumeUsage)
	if err != nil {
		return fmt.Errorf("unable to prune unused build cache: %w", err)
	}
	vu.UsedBytes -= reportBuildCache.SpaceReclaimed

	logboek.Context(ctx).Default().LogBlock("Prune unused build cache").Do(func() {
		logboek.Context(ctx).Default().LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
		logboek.Context(ctx).Default().LogF("Target volume usage percentage: %s > %s — %s\n", utils.RedF("%0.2f%%", vu.Percentage()), utils.BlueF("%d", targetVolumeUsage), utils.RedF("HIGH VOLUME USAGE"))
		logboek.Context(ctx).Default().LogF("Freed space: %s\n", utils.RedF("%s", humanize.Bytes(reportBuildCache.SpaceReclaimed)))
		logDeletedItems(ctx, reportBuildCache.ItemsDeleted)
	})

	// Step 2. Prune dangling images
	reportImages, err := cleaner.pruneImages(ctx, options, vu.UsedBytes, targetVolumeUsage)
	if err != nil {
		return fmt.Errorf("unable to prune dangling images: %w", err)
	}
	vu.UsedBytes -= reportImages.SpaceReclaimed

	logboek.Context(ctx).Default().LogBlock("Prune dangling images").Do(func() {
		logboek.Context(ctx).Default().LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
		logboek.Context(ctx).Default().LogF("Target volume usage percentage: %s > %s — %s\n", utils.RedF("%0.2f%%", vu.Percentage()), utils.BlueF("%d", targetVolumeUsage), utils.RedF("HIGH VOLUME USAGE"))
		logboek.Context(ctx).Default().LogF("Freed space: %s\n", utils.RedF("%s", humanize.Bytes(reportImages.SpaceReclaimed)))
		logDeletedItems(ctx, reportImages.ItemsDeleted)
	})

	// Step 3. Prune stopped containers
	reportContainers, err := cleaner.pruneContainers(ctx, options, vu.UsedBytes, targetVolumeUsage)
	if err != nil {
		return fmt.Errorf("unable to prune stopped containers: %w", err)
	}
	vu.UsedBytes -= reportContainers.SpaceReclaimed

	logboek.Context(ctx).Default().LogBlock("Prune stopped containers").Do(func() {
		logboek.Context(ctx).Default().LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
		logboek.Context(ctx).Default().LogF("Target volume usage percentage: %s > %s — %s\n", utils.RedF("%0.2f%%", vu.Percentage()), utils.BlueF("%d", targetVolumeUsage), utils.RedF("HIGH VOLUME USAGE"))
		logboek.Context(ctx).Default().LogF("Freed space: %s\n", utils.RedF("%s", humanize.Bytes(reportContainers.SpaceReclaimed)))
		logDeletedItems(ctx, reportContainers.ItemsDeleted)
	})

	// Step 4. Prune unused anonymous volumes
	reportVolumes, err := cleaner.pruneVolumes(ctx, options, vu.UsedBytes, targetVolumeUsage)
	if err != nil {
		return fmt.Errorf("unable to prune unused anonymous volumes: %w", err)
	}
	vu.UsedBytes -= reportVolumes.SpaceReclaimed

	logboek.Context(ctx).Default().LogBlock("Prune unused anonymous volumes").Do(func() {
		logboek.Context(ctx).Default().LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
		logboek.Context(ctx).Default().LogF("Target volume usage percentage: %s > %s — %s\n", utils.RedF("%0.2f%%", vu.Percentage()), utils.BlueF("%d", targetVolumeUsage), utils.RedF("HIGH VOLUME USAGE"))
		logboek.Context(ctx).Default().LogF("Freed space: %s\n", utils.RedF("%s", humanize.Bytes(reportVolumes.SpaceReclaimed)))
		logDeletedItems(ctx, reportVolumes.ItemsDeleted)
	})

	totalFreedBytes := reportBuildCache.SpaceReclaimed + reportImages.SpaceReclaimed + reportContainers.SpaceReclaimed + reportVolumes.SpaceReclaimed

	if vu.Percentage() <= options.AllowedStorageVolumeUsagePercentage {
		logboek.Context(ctx).Default().LogBlock("Local %s backend storage check", cleaner.BackendName()).Do(func() {
			logboek.Context(ctx).Default().LogF("Storage path: %s\n", backendStoragePath)
			logboek.Context(ctx).Default().LogF("Total freed space: %s\n", utils.RedF("%s", humanize.Bytes(totalFreedBytes)))
			logboek.Context(ctx).Default().LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
			logboek.Context(ctx).Default().LogF("Allowed percentage level exceeded: %s > %s — %s\n", utils.RedF("%0.2f%%", vu.Percentage()), utils.YellowF("%0.2f%%", options.AllowedStorageVolumeUsagePercentage), utils.RedF("HIGH VOLUME USAGE"))
			logboek.Context(ctx).Default().LogF("Target percentage level after cleanup: %0.2f%% - %0.2f%% (margin) = %s\n", options.AllowedStorageVolumeUsagePercentage, options.AllowedStorageVolumeUsageMarginPercentage, utils.BlueF("%0.2f%%", targetVolumeUsage))
		})

		return nil
	}

	// Step 5. Remove werf artifacts
	reportWerfArtifacts, err := cleaner.removeWerfArtifactsOld(ctx, options, vu.UsedBytes, targetVolumeUsage)
	if err != nil {
		return fmt.Errorf("unable to prune werf artifacts: %w", err)
	}
	vu.UsedBytes -= reportVolumes.SpaceReclaimed

	logboek.Context(ctx).Default().LogBlock("Remove werf images and containers", cleaner.BackendName()).Do(func() {
		logboek.Context(ctx).Default().LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
		logboek.Context(ctx).Default().LogF("Target volume usage percentage: %s > %s — %s\n", utils.RedF("%0.2f%%", vu.Percentage()), utils.BlueF("%d", targetVolumeUsage), utils.RedF("HIGH VOLUME USAGE"))
		logboek.Context(ctx).Default().LogF("Freed space: %s\n", utils.RedF("%s", humanize.Bytes(reportWerfArtifacts.SpaceReclaimed)))
		logDeletedItems(ctx, reportWerfArtifacts.ItemsDeleted)
	})

	if vu.UsedBytes > targetVolumeUsage {
		logboek.Context(ctx).Warn().LogOptionalLn()
		logboek.Context(ctx).Warn().LogF("WARNING: Detected high %s storage volume usage, while no werf images available to cleanup!\n", cleaner.BackendName())
		logboek.Context(ctx).Warn().LogF("WARNING:\n")
		logboek.Context(ctx).Warn().LogF("WARNING: Werf tries to maintain host clean by deleting:\n")
		logboek.Context(ctx).Warn().LogF("WARNING:  - old unused files from werf caches (which are stored in the ~/.werf/local_cache);\n")
		logboek.Context(ctx).Warn().LogF("WARNING:  - old temporary service files /tmp/werf-project-data-* and /tmp/werf-config-render-*;\n")
		logboek.Context(ctx).Warn().LogF("WARNING:  - least recently used werf images except local stages storage images (images built with 'werf build' without '--repo' param, or with '--stages-storage=:local' param for the werf v1.1).\n")
		logboek.Context(ctx).Warn().LogOptionalLn()
	}

	return nil
}

// pruneBuildCache removes all unused cache
func (cleaner *LocalBackendCleaner) pruneBuildCache(ctx context.Context, _ RunGCOptions, _, _ uint64) (prune.Report, error) {
	// TODO(a.zaytsev): handle dry-run mode
	return cleaner.backend.PruneBuildCache(ctx, prune.Options{})
}

// pruneContainers removes all stopped containers
func (cleaner *LocalBackendCleaner) pruneContainers(ctx context.Context, _ RunGCOptions, _, _ uint64) (prune.Report, error) {
	// TODO(a.zaytsev): handle dry-run mode
	return cleaner.backend.PruneContainers(ctx, prune.Options{})
}

// pruneImages removes all dangling images
func (cleaner *LocalBackendCleaner) pruneImages(ctx context.Context, _ RunGCOptions, _, _ uint64) (prune.Report, error) {
	// TODO(a.zaytsev): handle dry-run mode
	return cleaner.backend.PruneImages(ctx, prune.Options{})
}

// pruneVolumes removes all anonymous volumes not used by at least one container
func (cleaner *LocalBackendCleaner) pruneVolumes(ctx context.Context, _ RunGCOptions, _, _ uint64) (prune.Report, error) {
	// TODO(a.zaytsev): handle dry-run mode
	return cleaner.backend.PruneVolumes(ctx, prune.Options{})
}

// removeWerfArtifactsOld is old removal process implementation
func (cleaner *LocalBackendCleaner) removeWerfArtifactsOld(ctx context.Context, options RunGCOptions, actualVolumeUsage, targetVolumeUsage uint64) (prune.Report, error) {
	bytesToFree := targetVolumeUsage - actualVolumeUsage

	images, err := cleaner.safeWerfImages(ctx)
	if err != nil {
		return prune.Report{}, err
	}

	var report prune.Report
	var processedImagesIDs []string
	var processedContainersIDs []string

	for {
		var freedBytes uint64
		var freedImagesCount uint64
		var acquiredHostLocks []lockgate.LockHandle

		if len(images) > 0 {
			if err := logboek.Context(ctx).Default().LogProcess("Running cleanup for least recently used %s images created by werf", cleaner.BackendName()).DoError(func() error {
			DeleteImages:
				for _, imgSummary := range images {
					for _, id := range processedImagesIDs {
						if imgSummary.ID == id {
							logboek.Context(ctx).Default().LogFDetails("Skip already processed image %q\n", imgSummary.ID)
							continue DeleteImages
						}
					}
					processedImagesIDs = append(processedImagesIDs, imgSummary.ID)

					imageRemoved := false

					if len(imgSummary.RepoTags) > 0 {
						allTagsRemoved := true

						for _, ref := range imgSummary.RepoTags {
							if ref == "<none>:<none>" {
								if err := cleaner.removeImage(ctx, imgSummary.ID, options.Force, options.DryRun); err != nil {
									logboek.Context(ctx).Warn().LogF("failed to remove local %s image by ID %q: %s\n", cleaner.BackendName(), imgSummary.ID, err)
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
					} else if len(imgSummary.RepoDigests) > 0 {
						allDigestsRemoved := true

						for _, repoDigest := range imgSummary.RepoDigests {
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
						freedImgBytes := uint64(imgSummary.Size - normalizeSharedSize(imgSummary.SharedSize))

						freedBytes += freedImgBytes
						freedImagesCount++

						report.ItemsDeleted = append(report.ItemsDeleted, imgSummary.ID)
						report.SpaceReclaimed += freedImgBytes
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
				return report, err
			}
		}

		for _, lock := range acquiredHostLocks {
			if err := chart.ReleaseHostLock(lock); err != nil {
				return report, fmt.Errorf("unable to release lock %q: %w", lock.LockName, err)
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
			return report, err
		}

		if err := logboek.Context(ctx).Default().LogProcess("Running cleanup for dangling %s images created by werf", cleaner.BackendName()).DoError(func() error {
			return cleaner.safeDanglingImagesCleanup(ctx, commonOptions)
		}); err != nil {
			return report, err
		}

		if freedImagesCount == 0 {
			break
		}
		if options.DryRun {
			break
		}

		bytesToFree -= freedBytes

		if bytesToFree <= freedBytes {
			break
		}
	}

	return report, nil
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

func calculateBytesToFree(vu volumeutils.VolumeUsage, targetVolumeUsage uint64) uint64 {
	allowedVolumeUsageToFree := vu.Percentage() - float64(targetVolumeUsage)
	bytesToFree := uint64((float64(vu.TotalBytes) / 100.0) * allowedVolumeUsageToFree)
	return bytesToFree
}

func logDeletedItems(ctx context.Context, deletedItems []string) {
	for _, item := range deletedItems {
		logboek.Context(ctx).Default().LogLn(item)
	}
}
