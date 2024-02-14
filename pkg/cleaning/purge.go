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
	StorageManager                manager.StorageManagerInterface
	ProjectName                   string
	RmContainersThatUseWerfImages bool
	DryRun                        bool
}

func (m *purgeManager) run(ctx context.Context) error {
	if err := logboek.Context(ctx).Default().LogProcess("Deleting stages").DoError(func() error {
		stages, err := m.StorageManager.GetStageDescriptionListWithCache(ctx)
		if err != nil {
			return err
		}

		return m.deleteStages(ctx, stages, false)
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).Default().LogProcess("Deleting imports metadata").DoError(func() error {
		importMetadataIDs, err := m.StorageManager.GetStagesStorage().GetImportMetadataIDs(ctx, m.ProjectName, storage.WithCache())
		if err != nil {
			return err
		}

		return m.deleteImportsMetadata(ctx, importMetadataIDs)
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).Default().LogProcess("Deleting managed images").DoError(func() error {
		managedImages, err := m.StorageManager.GetStagesStorage().GetManagedImages(ctx, m.ProjectName, storage.WithCache())
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

	if err := m.purgeImageMetadata(ctx); err != nil {
		return err
	}

	if err := m.deleteCustomTags(ctx); err != nil {
		return err
	}

	if m.StorageManager.GetFinalStagesStorage() != nil {
		if err := logboek.Context(ctx).Default().LogProcess("Deleting final stages").DoError(func() error {
			finalStages, err := m.StorageManager.GetFinalStageDescriptionList(ctx)
			if err != nil {
				return err
			}

			return m.deleteStages(ctx, finalStages, true)
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

func (m *purgeManager) purgeImageMetadata(ctx context.Context) error {
	return purgeImageMetadata(ctx, m.ProjectName, m.StorageManager, m.DryRun)
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

		if err := deleteCustomTags(ctx, m.StorageManager, customTagList, m.DryRun); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func deleteCustomTags(ctx context.Context, storageManager manager.StorageManagerInterface, customTagList []string, dryRun bool) error {
	if dryRun {
		for _, customTag := range customTagList {
			logboek.Context(ctx).Default().LogFDetails("  tag: %s\n", customTag)
			logboek.Context(ctx).Default().LogOptionalLn()
		}

		return nil
	}

	if err := storageManager.ForEachDeleteStageCustomTag(ctx, customTagList, func(ctx context.Context, tag string, err error) error {
		if err != nil {
			if err := handleDeletionError(err); err != nil {
				return err
			}
		}

		logboek.Context(ctx).Default().LogFDetails("  tag: %s\n", tag)

		return nil
	}); err != nil {
		return err
	}

	return nil
}
