package cleaning

import (
	"context"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/cleaning/stage_manager"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/logging"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/storage/manager"
)

type PurgeOptions struct {
	RmContainersThatUseWerfImages bool
	DryRun                        bool
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
	}
}

type purgeManager struct {
	StorageManager                *manager.StorageManager
	ProjectName                   string
	RmContainersThatUseWerfImages bool
	DryRun                        bool
}

func (m *purgeManager) run(ctx context.Context) error {
	if err := logboek.Context(ctx).Default().LogProcess("Deleting stages").DoError(func() error {
		stages, err := m.StorageManager.GetStageDescriptionList(ctx)
		if err != nil {
			return err
		}

		return m.deleteStages(ctx, stages, false)
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).Default().LogProcess("Deleting imports metadata").DoError(func() error {
		importMetadataIDs, err := m.StorageManager.GetStagesStorage().GetImportMetadataIDs(ctx, m.ProjectName)
		if err != nil {
			return err
		}

		return m.deleteImportsMetadata(ctx, importMetadataIDs)
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).Default().LogProcess("Deleting managed images").DoError(func() error {
		managedImages, err := m.StorageManager.GetStagesStorage().GetManagedImages(ctx, m.ProjectName)
		if err != nil {
			return err
		}

		if err := m.deleteManagedImages(ctx, managedImages); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).Default().LogProcess("Deleting images metadata").DoError(func() error {
		_, imageMetadataByImageName, err := m.StorageManager.GetStagesStorage().GetAllAndGroupImageMetadataByImageName(ctx, m.ProjectName, []string{})
		if err != nil {
			return err
		}

		for imageNameID, stageIDCommitList := range imageMetadataByImageName {
			if err := m.deleteImageMetadata(ctx, imageNameID, stageIDCommitList); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	if err := m.deleteCustomTags(ctx); err != nil {
		return err
	}

	if m.StorageManager.GetFinalStagesStorage() != nil {
		if err := logboek.Context(ctx).Default().LogProcess("Deleting final stages").DoError(func() error {
			stages, err := m.StorageManager.GetStageDescriptionList(ctx)
			if err != nil {
				return err
			}

			return m.deleteStages(ctx, stages, true)
		}); err != nil {
			return err
		}
	}

	return nil
}

func (m *purgeManager) deleteStages(ctx context.Context, stages []*image.StageDescription, isFinal bool) error {
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

	return deleteStages(ctx, m.StorageManager, m.DryRun, deleteStageOptions, stages, isFinal)
}

func (m *purgeManager) deleteImportsMetadata(ctx context.Context, importsMetadataIDs []string) error {
	return deleteImportsMetadata(ctx, m.ProjectName, m.StorageManager, importsMetadataIDs, m.DryRun)
}

func (m *purgeManager) deleteManagedImages(ctx context.Context, managedImages []string) error {
	if m.DryRun {
		for _, managedImage := range managedImages {
			logboek.Context(ctx).Default().LogFDetails("  name: %s\n", logging.ImageLogName(managedImage, false))
			logboek.Context(ctx).LogOptionalLn()
		}
		return nil
	}

	return m.StorageManager.ForEachRmManagedImage(ctx, m.ProjectName, managedImages, func(ctx context.Context, managedImage string, err error) error {
		if err != nil {
			if err := handleDeletionError(err); err != nil {
				return err
			}

			logboek.Context(ctx).Warn().LogF("WARNING: Managed image %s deletion failed: %s\n", managedImage, err)

			return nil
		}

		logboek.Context(ctx).Default().LogFDetails("  name: %s\n", logging.ImageLogName(managedImage, false))

		return nil
	})
}

func (m *purgeManager) deleteImageMetadata(ctx context.Context, imageNameOrID string, stageIDCommitList map[string][]string) error {
	return deleteImageMetadata(ctx, m.ProjectName, m.StorageManager, imageNameOrID, stageIDCommitList, m.DryRun)
}

func (m *purgeManager) deleteCustomTags(ctx context.Context) error {
	stageIDCustomTagList, err := stage_manager.GetCustomTagsMetadata(ctx, m.StorageManager)
	if err != nil {
		return err
	}

	if err := logboek.Context(ctx).LogProcess("Deleting custom tags").DoError(func() error {
		for _, customTagList := range stageIDCustomTagList {
			for _, customTag := range customTagList {
				if err := deleteCustomTag(ctx, m.StorageManager.GetStagesStorage(), customTag); err != nil {
					return err
				}
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func deleteCustomTag(ctx context.Context, stagesStorage storage.StagesStorage, customTag string) error {
	err := stagesStorage.DeleteStageCustomTag(ctx, customTag)
	if err != nil {
		if err := handleDeletionError(err); err != nil {
			return err
		}

		return nil
	}

	logboek.Context(ctx).Default().LogFDetails("  tag: %s\n", customTag)
	logboek.Context(ctx).Default().LogOptionalLn()

	return nil
}
