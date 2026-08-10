package cleaning

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/cleaning/stage_manager"
	"github.com/werf/werf/v2/pkg/cleanup_report"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/manager"
)

type PurgeOptions struct {
	RmContainersThatUseWerfImages bool
	DryRun                        bool
	Report                        *cleanup_report.Report
}

func Purge(ctx context.Context, projectName string, storageManager *manager.StorageManager, options PurgeOptions) error {
	return newPurgeManager(projectName, storageManager, options).run(ctx)
}

func newPurgeManager(projectName string, storageManager *manager.StorageManager, options PurgeOptions) *purgeManager {
	return &purgeManager{
		StorageManager:                storageManager,
		ProjectName:                   projectName,
		RmContainersThatUseWerfImages: options.RmContainersThatUseWerfImages,
		DryRun:                        options.DryRun,
		report:                        options.Report,
	}
}

type purgeManager struct {
	StorageManager                manager.StorageManagerInterface
	ProjectName                   string
	RmContainersThatUseWerfImages bool
	DryRun                        bool

	report *cleanup_report.Report
}

func (m *purgeManager) run(ctx context.Context) error {
	if err := logboek.Context(ctx).Default().LogProcess("Deleting stages").DoError(func() error {
		stageDescSet, err := m.StorageManager.GetStageDescSetWithCache(ctx)
		if err != nil {
			return err
		}

		return m.deleteStageDescSet(ctx, stageDescSet, false)
	}); err != nil {
		return err
	}

	if err := m.purgeManagedImages(ctx); err != nil {
		return err
	}

	if err := m.purgeImageMetadata(ctx); err != nil {
		return err
	}

	if err := m.deleteCustomTags(ctx); err != nil {
		return err
	}

	if err := logboek.Context(ctx).Default().LogProcess("Deleting rejected stages").DoError(func() error {
		return m.purgeRejectedStages(ctx)
	}); err != nil {
		return err
	}

	if m.StorageManager.GetFinalStagesStorage() != nil {
		if err := logboek.Context(ctx).Default().LogProcess("Deleting final stages").DoError(func() error {
			finalDescSet, err := m.StorageManager.GetFinalStageDescSet(ctx)
			if err != nil {
				return err
			}

			return m.deleteStageDescSet(ctx, finalDescSet, true)
		}); err != nil {
			return err
		}
	}

	return m.deleteMetaRepoMarker(ctx)
}

// deleteMetaRepoMarker runs last, so a purge that failed to delete some metadata
// never leaves the meta-repo populated with the safeguard already gone.
func (m *purgeManager) deleteMetaRepoMarker(ctx context.Context) error {
	repo, ok := m.StorageManager.GetStagesStorage().(*storage.RepoStagesStorage)
	if !ok {
		return nil
	}

	metaRepoAddress, found, err := repo.GetMetaRepoMarker(ctx, m.ProjectName)
	if err != nil {
		return fmt.Errorf("unable to get meta-repo marker: %w", err)
	}
	if !found {
		return nil
	}

	return logboek.Context(ctx).Default().LogProcess("Deleting meta-repo safeguard marker").DoError(func() error {
		logboek.Context(ctx).Default().LogFWithCustomStyle(deletedStyle, "  meta-repo: %s\n", metaRepoAddress)

		if m.DryRun {
			return nil
		}

		if err := repo.RmMetaRepoMarker(ctx, m.ProjectName); err != nil {
			return fmt.Errorf("unable to remove meta-repo marker: %w", err)
		}

		return nil
	})
}

func (m *purgeManager) deleteStageDescSet(ctx context.Context, stageDescSet image.StageDescSet, isFinal bool) error {
	deleteStageOptions := manager.ForEachDeleteStageOptions{
		DeleteImageOptions: storage.DeleteImageOptions{
			RmiForce: true,
		},
		FilterStagesAndProcessRelatedDataOptions: storage.FilterStagesAndProcessRelatedDataOptions{
			SkipUsedImage:            false,
			RmForce:                  m.RmContainersThatUseWerfImages,
			RmContainersThatUseImage: m.RmContainersThatUseWerfImages,
		},
	}

	return deleteStageDescSet(ctx, m.StorageManager, m.DryRun, deleteStageOptions, stageDescSet, isFinal, m.report)
}

func (m *purgeManager) purgeImageMetadata(ctx context.Context) error {
	return purgeImageMetadata(ctx, m.ProjectName, m.StorageManager, m.DryRun, m.report)
}

func (m *purgeManager) purgeManagedImages(ctx context.Context) error {
	return purgeManagedImages(ctx, m.ProjectName, m.StorageManager, m.DryRun, m.report)
}

func (m *purgeManager) deleteCustomTags(ctx context.Context) error {
	stageIDCustomTagList, err := stage_manager.GetCustomTagsMetadata(ctx, m.StorageManager)
	if err != nil {
		return err
	}

	if err := logboek.Context(ctx).LogProcess("Deleting custom tags").DoError(func() error {
		var customTagList []string
		for _, list := range stageIDCustomTagList {
			customTagList = append(customTagList, list...)
		}

		if err := deleteCustomTags(ctx, m.StorageManager, customTagList, m.DryRun, m.report); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func deleteCustomTags(ctx context.Context, storageManager manager.StorageManagerInterface, customTagList []string, dryRun bool, report *cleanup_report.Report) error {
	if dryRun {
		for _, customTag := range customTagList {
			logboek.Context(ctx).Default().LogFWithCustomStyle(deletedStyle, "  tag: %s\n", customTag)
			logboek.Context(ctx).Default().LogOptionalLn()
			report.AddDeleted(ctx, cleanup_report.Item{Type: cleanup_report.ItemTypeCustomTag, Tag: customTag})
		}

		return nil
	}

	if err := storageManager.ForEachDeleteStageCustomTag(ctx, customTagList, func(ctx context.Context, tag string, err error) error {
		if err != nil {
			if err := handleDeletionError(err); err != nil {
				return err
			}
		} else {
			report.AddDeleted(ctx, cleanup_report.Item{Type: cleanup_report.ItemTypeCustomTag, Tag: tag})
		}

		logboek.Context(ctx).Default().LogFWithCustomStyle(deletedStyle, "  tag: %s\n", tag)

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (m *purgeManager) purgeRejectedStages(ctx context.Context) error {
	customTagsByStageID, err := stage_manager.GetCustomTagsMetadata(ctx, m.StorageManager)
	if err != nil {
		return fmt.Errorf("unable to get custom tags metadata: %w", err)
	}

	if _, err := deleteRejectedStagesWithLinkedTags(ctx, m.StorageManager, customTagsByStageID, m.DryRun, m.report); err != nil {
		return err
	}
	return nil
}
