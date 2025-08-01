package manager

import (
	"context"
	"fmt"

	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/lrumeta"
)

type CopyStageOptions struct {
	ContainerBackend     container_backend.ContainerBackend
	LegacyImage          container_backend.LegacyImageInterface
	FetchStage           stage.Interface
	IsMultiplatformImage bool
}

func (m *StorageManager) CopyStage(ctx context.Context, src, dest storage.StagesStorage, stageID image.StageID, opts CopyStageOptions) (*image.StageDesc, error) {
	switch typedSrc := src.(type) {
	case *storage.LocalStagesStorage:
		return m.copyStageFromLocalStorage(ctx, typedSrc, dest, stageID, opts)
	case *storage.RepoStagesStorage:
		return dest.CopyFromStorage(ctx, src, m.ProjectName, stageID, storage.CopyFromStorageOptions{IsMultiplatformImage: opts.IsMultiplatformImage})
	default:
		panic(fmt.Sprintf("not implemented for storage %s", typedSrc))
	}
}

func (m *StorageManager) copyStageFromLocalStorage(ctx context.Context, src *storage.LocalStagesStorage, dest storage.StagesStorage, stageID image.StageID, opts CopyStageOptions) (*image.StageDesc, error) {
	if opts.LegacyImage == nil {
		panic("expected non empty LegacyImage parameter")
	}
	if opts.ContainerBackend == nil {
		panic("expected non empty ContainerBackend parameter")
	}

	if opts.FetchStage != nil {
		if _, err := m.FetchStage(ctx, opts.ContainerBackend, opts.FetchStage); err != nil {
			return nil, fmt.Errorf("unable to fetch stage %s: %w", opts.FetchStage.LogDetailedName(), err)
		}
	}

	newImg := opts.LegacyImage.GetCopy()
	destImageName := dest.ConstructStageImageName(m.ProjectName, stageID.Digest, stageID.CreationTs)

	if err := opts.ContainerBackend.RenameImage(ctx, newImg, destImageName, false); err != nil {
		return nil, fmt.Errorf("unable to rename image %s to %s: %w", opts.LegacyImage.Name(), destImageName, err)
	}
	if err := dest.StoreImage(ctx, newImg); err != nil {
		return nil, fmt.Errorf("unable to store stage %s into the stages storage %s: %w", stageID.String(), dest.String(), err)
	}

	if err := storeStageDescIntoLocalManifestCache(ctx, m.ProjectName, stageID, dest, ConvertStageDescForStagesStorage(newImg.GetStageDesc(), dest)); err != nil {
		return nil, fmt.Errorf("error storing stage %s description into local manifest cache: %w", destImageName, err)
	}
	if err := lrumeta.CommonLRUImagesCache.AccessImage(ctx, destImageName); err != nil {
		return nil, fmt.Errorf("error accessing last recently used images cache for %s: %w", destImageName, err)
	}
	return newImg.GetStageDesc(), nil
}
