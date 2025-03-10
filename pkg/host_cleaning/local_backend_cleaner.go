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

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/kubedog/pkg/utils"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/container_backend/prune"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage/lrumeta"
	"github.com/werf/werf/v2/pkg/volumeutils"
)

var ErrUnsupportedContainerBackend = errors.New("unsupported container backend")

var errOptionDryRunNotSupported = errors.New("option dry-run not supported")

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

type cleanupReport prune.Report

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

// werfImages returns werf images are safe for removing
func (cleaner *LocalBackendCleaner) werfImages(ctx context.Context) (image.ImagesList, error) {
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

	// We can clarify StoragePath from now to further usage
	options.StoragePath = backendStoragePath

	logboek.Context(ctx).LogF("Storage path: %s\n", options.StoragePath)
	logboek.Context(ctx).LogOptionalLn()

	vu, err := cleaner.volumeutilsGetVolumeUsageByPath(ctx, options.StoragePath)
	if err != nil {
		return fmt.Errorf("error getting volume usage by path %q: %w", options.StoragePath, err)
	}

	if vu.Percentage() <= options.AllowedStorageVolumeUsagePercentage {
		logboek.Context(ctx).LogBlock("Check storage").Do(func() {
			logboek.Context(ctx).LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
			logboek.Context(ctx).LogF("Allowed volume usage percentage: %s <= %s — %s\n", utils.GreenF("%0.2f%%", vu.Percentage()), utils.BlueF("%0.2f%%", options.AllowedStorageVolumeUsagePercentage), utils.GreenF("OK"))
		})

		return nil
	}

	targetVolumeUsagePercentage := math.Max(options.AllowedStorageVolumeUsagePercentage-options.AllowedStorageVolumeUsageMarginPercentage, 0)

	logboek.Context(ctx).LogBlock("Check storage").Do(func() {
		logboek.Context(ctx).LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
		logboek.Context(ctx).LogF("Allowed percentage level exceeded: %s > %s — %s\n", utils.RedF("%0.2f%%", vu.Percentage()), utils.YellowF("%0.2f%%", options.AllowedStorageVolumeUsagePercentage), utils.RedF("HIGH VOLUME USAGE"))
		logboek.Context(ctx).LogF("Target percentage level after cleanup: %0.2f%% - %0.2f%% (margin) = %s\n", options.AllowedStorageVolumeUsagePercentage, options.AllowedStorageVolumeUsageMarginPercentage, utils.BlueF("%d", targetVolumeUsagePercentage))
		logboek.Context(ctx).LogF("Needed to free: %s\n", utils.RedF("%s", humanize.Bytes(calculateBytesToFree(vu, targetVolumeUsagePercentage))))
	})

	vuBefore := vu

	// Step 1. Prune unused build cache
	err = logboek.Context(ctx).LogBlock("Prune unused build cache").DoError(func() error {
		reportBuildCache, err := cleaner.pruneBuildCache(ctx, options)
		if handleError(ctx, err) != nil {
			return err
		}

		vu.UsedBytes -= reportBuildCache.SpaceReclaimed

		logboek.Context(ctx).LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
		logboek.Context(ctx).LogF("Target volume usage percentage: %s > %s — %s\n", utils.RedF("%0.2f%%", vu.Percentage()), utils.BlueF("%d", targetVolumeUsagePercentage), utils.RedF("HIGH VOLUME USAGE"))
		logboek.Context(ctx).LogF("Freed space: %s\n", utils.RedF("%s", humanize.Bytes(reportBuildCache.SpaceReclaimed)))
		logDeletedItems(ctx, reportBuildCache.ItemsDeleted)

		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to prune unused build cache: %w", err)
	}

	// Step 2. Prune stopped containers
	err = logboek.Context(ctx).LogBlock("Prune stopped containers").DoError(func() error {
		reportContainers, err := cleaner.pruneContainers(ctx, options)
		if handleError(ctx, err) != nil {
			return err
		}

		vu.UsedBytes -= reportContainers.SpaceReclaimed

		logboek.Context(ctx).LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
		logboek.Context(ctx).LogF("Target volume usage percentage: %s > %s — %s\n", utils.RedF("%0.2f%%", vu.Percentage()), utils.BlueF("%d", targetVolumeUsagePercentage), utils.RedF("HIGH VOLUME USAGE"))
		logboek.Context(ctx).LogF("Freed space: %s\n", utils.RedF("%s", humanize.Bytes(reportContainers.SpaceReclaimed)))
		logDeletedItems(ctx, reportContainers.ItemsDeleted)

		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to prune stopped containers: %w", err)
	}

	// Step 3. Prune unused anonymous volumes
	err = logboek.Context(ctx).LogBlock("Prune unused anonymous volumes").DoError(func() error {
		reportVolumes, err := cleaner.pruneVolumes(ctx, options)
		if handleError(ctx, err) != nil {
			return err
		}

		vu.UsedBytes -= reportVolumes.SpaceReclaimed

		logboek.Context(ctx).LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
		logboek.Context(ctx).LogF("Target volume usage percentage: %s > %s — %s\n", utils.RedF("%0.2f%%", vu.Percentage()), utils.BlueF("%d", targetVolumeUsagePercentage), utils.RedF("HIGH VOLUME USAGE"))
		logboek.Context(ctx).LogF("Freed space: %s\n", utils.RedF("%s", humanize.Bytes(reportVolumes.SpaceReclaimed)))
		logDeletedItems(ctx, reportVolumes.ItemsDeleted)

		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to prune unused anonymous volumes: %w", err)
	}

	// Step 4. Prune dangling images
	err = logboek.Context(ctx).LogBlock("Prune dangling images").DoError(func() error {
		reportImages, err := cleaner.pruneImages(ctx, options)
		if handleError(ctx, err) != nil {
			return err
		}

		vu.UsedBytes -= reportImages.SpaceReclaimed

		logboek.Context(ctx).LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
		logboek.Context(ctx).LogF("Target volume usage percentage: %s > %s — %s\n", utils.RedF("%0.2f%%", vu.Percentage()), utils.BlueF("%d", targetVolumeUsagePercentage), utils.RedF("HIGH VOLUME USAGE"))
		logboek.Context(ctx).LogF("Freed space: %s\n", utils.RedF("%s", humanize.Bytes(reportImages.SpaceReclaimed)))
		logDeletedItems(ctx, reportImages.ItemsDeleted)

		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to prune dangling images: %w", err)
	}

	if vu.Percentage() <= options.AllowedStorageVolumeUsagePercentage {
		freedBytes := vuBefore.UsedBytes - vu.UsedBytes

		logboek.Context(ctx).LogBlock("Check storage", cleaner.BackendName()).Do(func() {
			logboek.Context(ctx).LogF("Total freed space: %s\n", utils.RedF("%s", humanize.Bytes(freedBytes)))
			logboek.Context(ctx).LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
			logboek.Context(ctx).LogF("Allowed percentage level exceeded: %s > %s — %s\n", utils.RedF("%0.2f%%", vu.Percentage()), utils.YellowF("%0.2f%%", options.AllowedStorageVolumeUsagePercentage), utils.RedF("HIGH VOLUME USAGE"))
			logboek.Context(ctx).LogF("Target percentage level after cleanup: %0.2f%% - %0.2f%% (margin) = %s\n", options.AllowedStorageVolumeUsagePercentage, options.AllowedStorageVolumeUsageMarginPercentage, utils.BlueF("%0.2f%%", targetVolumeUsagePercentage))
		})

		return nil
	}

	// Step 5. Remove werf containers
	err = logboek.Context(ctx).LogBlock("Cleanup werf containers").DoError(func() error {
		vuBefore = vu

		reportWerfContainers, err := cleaner.safeCleanupWerfContainers(ctx, options)
		if err != nil {
			return err
		}

		// No explicit way to calculate reclaimed containers' size at all.
		// But we can calculate it implicitly via disk usage check.
		vu, err = cleaner.volumeutilsGetVolumeUsageByPath(ctx, options.StoragePath)
		if err != nil {
			return fmt.Errorf("error getting volume usage by path %q: %w", options.StoragePath, err)
		}

		freedBytes := vuBefore.UsedBytes - vu.UsedBytes

		logboek.Context(ctx).LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
		logboek.Context(ctx).LogF("Target volume usage percentage: %s > %s — %s\n", utils.RedF("%0.2f%%", vu.Percentage()), utils.BlueF("%d", targetVolumeUsagePercentage), utils.RedF("HIGH VOLUME USAGE"))
		logboek.Context(ctx).LogF("Freed space: %s\n", utils.RedF("%s", humanize.Bytes(freedBytes)))

		logDeletedItems(ctx, reportWerfContainers.ItemsDeleted)

		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to remove werf containers: %w", err)
	}

	// Step 6. Remove werf images
	err = logboek.Context(ctx).LogBlock("Cleanup werf images").DoError(func() error {
		reportWerfImages, err := cleaner.safeCleanupWerfImages(ctx, options, vu, targetVolumeUsagePercentage)
		if err != nil {
			return err
		}

		vu.UsedBytes -= reportWerfImages.SpaceReclaimed

		logboek.Context(ctx).LogF("Volume usage: %s / %s\n", humanize.Bytes(vu.UsedBytes), humanize.Bytes(vu.TotalBytes))
		logboek.Context(ctx).LogF("Target volume usage percentage: %s > %s — %s\n", utils.RedF("%0.2f%%", vu.Percentage()), utils.BlueF("%d", targetVolumeUsagePercentage), utils.RedF("HIGH VOLUME USAGE"))
		logboek.Context(ctx).LogF("Freed space: %s\n", utils.RedF("%s", humanize.Bytes(reportWerfImages.SpaceReclaimed)))
		logDeletedItems(ctx, reportWerfImages.ItemsDeleted)

		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to cleanup werf images: %w", err)
	}

	if vu.Percentage() > targetVolumeUsagePercentage {
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
func (cleaner *LocalBackendCleaner) pruneBuildCache(ctx context.Context, options RunGCOptions) (cleanupReport, error) {
	if options.DryRun {
		// NOTE: Buildah does not give us a way to precalculate pruned size.
		// NOTE: Docker does not give us a way to precalculate pruned size.
		return cleanupReport{}, errOptionDryRunNotSupported
	}

	report, err := cleaner.backend.PruneBuildCache(ctx, prune.Options{})
	return cleanupReport(report), err
}

// pruneContainers removes all stopped containers
func (cleaner *LocalBackendCleaner) pruneContainers(ctx context.Context, options RunGCOptions) (cleanupReport, error) {
	if options.DryRun {
		// NOTE: Buildah does not give us a way to precalculate pruned size.

		// NOTE: Docker give us an ability to precalculate container.Size, however it is expensive operation.
		// In docker case we could list containers using: status=exited AND size=true AND all=true:
		// https://pkg.go.dev/github.com/docker/docker/client@v25.0.5+incompatible#ContainerAPIClient.ContainerList

		return cleanupReport{}, errOptionDryRunNotSupported
	}

	report, err := cleaner.backend.PruneContainers(ctx, prune.Options{})
	return cleanupReport(report), err
}

// pruneImages removes all dangling images
func (cleaner *LocalBackendCleaner) pruneImages(ctx context.Context, options RunGCOptions) (cleanupReport, error) {
	if options.DryRun {
		list, err := cleaner.backend.Images(ctx, buildImagesOptions(
			util.NewPair("dangling", "true"),
		))
		if err != nil {
			return cleanupReport{}, err
		}
		return mapImageListToCleanupReport(list), nil
	}

	report, err := cleaner.backend.PruneImages(ctx, prune.Options{})
	return cleanupReport(report), err
}

// pruneVolumes removes all anonymous volumes not used by at least one container
func (cleaner *LocalBackendCleaner) pruneVolumes(ctx context.Context, options RunGCOptions) (cleanupReport, error) {
	if options.DryRun {
		// NOTE: Buildah does not give us a way to precalculate pruned size.
		// NOTE: Docker does not give us a way to precalculate pruned size.
		return cleanupReport{}, errOptionDryRunNotSupported
	}

	report, err := cleaner.backend.PruneVolumes(ctx, prune.Options{})
	return cleanupReport(report), err
}

// safeCleanupWerfContainers cleanups werf containers safely using host locks
func (cleaner *LocalBackendCleaner) safeCleanupWerfContainers(ctx context.Context, options RunGCOptions) (cleanupReport, error) {
	// TODO(a.zaytsev): add werf build containers
	containers, err := werfContainersByContainersOptions(ctx, cleaner.backend, buildContainersOptions())
	if err != nil {
		return cleanupReport{}, fmt.Errorf("cannot get build containers: %w", err)
	}

	if options.DryRun {
		return mapContainerListToCleanupReport(containers), nil
	}

	err = cleaner.doSafeCleanupWerfContainers(ctx, options, containers)
	if err != nil {
		return cleanupReport{}, err
	}

	return mapContainerListToCleanupReport(containers), nil
}

func (cleaner *LocalBackendCleaner) doSafeCleanupWerfContainers(ctx context.Context, options RunGCOptions, containers image.ContainerList) error {
	for _, container := range containers {
		containerName := werfContainerName(container)

		if containerName == "" {
			logboek.Context(ctx).Warn().LogF("Ignore bad container %s\n", container.ID)
			continue
		}

		err := withHostLock(ctx, container_backend.ContainerLockName(containerName), func() error {
			err := cleaner.backend.Rm(ctx, container.ID, container_backend.RmOpts{
				Force: options.Force,
			})
			if err != nil {
				return fmt.Errorf("failed to remove container %s: %w", logContainerName(container), err)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// safeCleanupWerfImages cleanups werf images safely using host locks
func (cleaner *LocalBackendCleaner) safeCleanupWerfImages(ctx context.Context, options RunGCOptions, vu volumeutils.VolumeUsage, targetVolumeUsagePercentage float64) (cleanupReport, error) {
	images, err := cleaner.werfImages(ctx)
	if err != nil {
		return cleanupReport{}, err
	}

	if options.DryRun {
		n := calculateImagesCountToFree(images, 0, vu, targetVolumeUsagePercentage)
		return mapImageListToCleanupReport(images[:n]), nil
	}

	report, err := cleaner.doSafeCleanupWerfImages(ctx, options, vu, targetVolumeUsagePercentage, images)
	if err != nil {
		return cleanupReport{}, err
	}

	return report, nil
}

func (cleaner *LocalBackendCleaner) doSafeCleanupWerfImages(ctx context.Context, options RunGCOptions, vu volumeutils.VolumeUsage, targetVolumeUsagePercentage float64, images image.ImagesList) (cleanupReport, error) {
	calcN := calculateImagesCountToFree
	tVu := targetVolumeUsagePercentage

	report := cleanupReport{
		ItemsDeleted:   make([]string, 0, len(images)),
		SpaceReclaimed: 0,
	}

	for i, n := 0, calcN(images, 0, vu, tVu); i < n; i, n = n, calcN(images, n, vu, tVu) {
		for _, imgSummary := range images[i:n] {

			var ok bool
			var err error

			if len(imgSummary.RepoTags) > 0 {
				ok, err = cleaner.removeImageByRepoTags(ctx, options, imgSummary)
				if err != nil {
					return report, err
				}
			} else {
				ok, err = cleaner.removeImageByRepoDigests(ctx, options, imgSummary)
			}

			if ok {
				report.ItemsDeleted = append(report.ItemsDeleted, imgSummary.ID)
			}
		}

		// Here we actually don't know how much space we reclaimed. There are several reasons of why:
		//	1. Image can have "shared size" but backend cannot calculate the shared size properly.
		//  2. An error can happen while removing of img.RepoTags of img.RepoDigests.
		//
		// So there is only one way to handle this is re-calculate of disk usage size (expensive operation):
		vuAfter, err := volumeutils.GetVolumeUsageByPath(ctx, options.StoragePath)
		if err != nil {
			return report, fmt.Errorf("error getting volume usage by path %q: %w", options.StoragePath, err)
		}

		report.SpaceReclaimed += vu.UsedBytes - vuAfter.UsedBytes
		// we must update vu variable to re-calculate how many images need to clean-up
		vu = vuAfter
	}

	return report, nil
}

func (cleaner *LocalBackendCleaner) removeImageByRepoTags(ctx context.Context, options RunGCOptions, imgSummary image.Summary) (bool, error) {
	unRemovedCount := 0

	for _, ref := range imgSummary.RepoTags {
		if ref == "<none>:<none>" { // an intermediate image
			err := cleaner.backend.Rmi(ctx, ref, container_backend.RmiOpts{
				Force: options.Force,
			})
			if err != nil {
				logboek.Context(ctx).Warn().LogF("failed to remove local image by ID %q: %s\n", imgSummary.ID, err)
				unRemovedCount++
			}
		} else {
			_ = withHostLock(ctx, container_backend.ImageLockName(ref), func() error {
				err := cleaner.backend.Rmi(ctx, ref, container_backend.RmiOpts{
					Force: options.Force,
				})
				if err != nil {
					logboek.Context(ctx).Warn().LogF("failed to remove local image by repo tag %q: %s\n", ref, err)
					unRemovedCount++
				}
				return nil
			})
		}
	}

	return unRemovedCount == 0, nil
}

func (cleaner *LocalBackendCleaner) removeImageByRepoDigests(ctx context.Context, options RunGCOptions, imgSummary image.Summary) (bool, error) {
	unRemovedCount := 0

	for _, repoDigest := range imgSummary.RepoDigests {
		err := cleaner.backend.Rmi(ctx, repoDigest, container_backend.RmiOpts{
			Force: options.Force,
		})
		if err != nil {
			logboek.Context(ctx).Warn().LogF("failed to remove local image by repo digest %q: %s\n", repoDigest, err)
			unRemovedCount++
		}
	}

	return unRemovedCount == 0, nil
}

func calculateBytesToFree(vu volumeutils.VolumeUsage, targetVolumeUsage float64) uint64 {
	allowedVolumeUsageToFree := math.Max(vu.Percentage()-targetVolumeUsage, 0)
	bytesToFree := uint64((float64(vu.TotalBytes) / 100.0) * allowedVolumeUsageToFree)
	return bytesToFree
}

func logDeletedItems(ctx context.Context, deletedItems []string) {
	for _, item := range deletedItems {
		logboek.Context(ctx).LogLn(item)
	}
}

func mapImageListToCleanupReport(list image.ImagesList) cleanupReport {
	report := cleanupReport{
		ItemsDeleted:   make([]string, 0, len(list)),
		SpaceReclaimed: 0,
	}
	for _, img := range list {
		report.ItemsDeleted = append(report.ItemsDeleted, img.ID)
		report.SpaceReclaimed += uint64(img.Size)
	}
	return report
}

// handleError is common error handler
func handleError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, container_backend.ErrUnsupportedFeature):
		logboek.Context(ctx).Warn().LogLn("Backend does not support this feature")
		return nil
	case errors.Is(err, errOptionDryRunNotSupported):
		logboek.Context(ctx).Warn().LogLn("There is not an able to precalculate size in --dry-run mode")
		return nil
	case err != nil:
		return err
	}
	return nil
}

func mapContainerListToCleanupReport(list image.ContainerList) cleanupReport {
	report := cleanupReport{
		ItemsDeleted:   make([]string, 0, len(list)),
		SpaceReclaimed: 0,
	}
	for _, container := range list {
		report.ItemsDeleted = append(report.ItemsDeleted, container.ID)
	}
	return report
}

// calculateImagesCountToFree returns index ∈ [0, len(images)] derived from actual and target volume usage
func calculateImagesCountToFree(images image.ImagesList, startIndex int, vu volumeutils.VolumeUsage, targetVolumeUsagePercentage float64) int {
	bytesToFree := calculateBytesToFree(vu, targetVolumeUsagePercentage)

	i := startIndex
	var freedBytes uint64

	for ; i < len(images) && freedBytes < bytesToFree; i++ {
		freedBytes += uint64(images[i].Size)
	}

	return i
}
