package build

import (
	"context"
	"fmt"
	"slices"

	"github.com/samber/lo"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/container_backend/filter"
	"github.com/werf/werf/v2/pkg/container_backend/label"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/sbom"
	"github.com/werf/werf/v2/pkg/sbom/scanner"
	"github.com/werf/werf/v2/pkg/storage"
)

type sbomPhase struct {
	containerBackend container_backend.ContainerBackend
	stagesStorage    storage.StagesStorage
	isLocalStorage   bool
}

func newSbomPhase(
	backend container_backend.ContainerBackend,
	stagesStorage storage.StagesStorage,
) *sbomPhase {
	_, isLocalStorage := stagesStorage.(*storage.LocalStagesStorage)

	return &sbomPhase{
		containerBackend: backend,
		stagesStorage:    stagesStorage,
		isLocalStorage:   isLocalStorage,
	}
}

// Converge searches relevant SBOM image in local and remote storages.
// If the relevant image is found, it does nothing.
// Otherwise, it generates new sbom image and pushes that image into remote storage.
func (phase *sbomPhase) Converge(ctx context.Context, stageDesc *image.StageDesc, scanOpts scanner.ScanOptions) error {
	sourceImageName := stageDesc.Info.Name
	sbomImageName := sbom.ImageName(sourceImageName)

	scanOpts.Commands[0].SourcePath = sourceImageName

	sbomImgLabels := phase.prepareSbomLabels(ctx, stageDesc.Info.Labels, scanOpts)

	return logboek.Context(ctx).Default().LogProcess("SBOM processing").DoError(func() error {
		_, ok, err := phase.findSbomImageLocally(ctx, sbomImgLabels, sbomImageName)
		if err != nil {
			return err
		}
		logboek.Context(ctx).Debug().LogF("-- sbom_phase.Converge: sbom image is found locally=%t\n", ok)

		if phase.isLocalStorage {
			if ok {
				return nil
			}
		} else {
			if ok {
				if _, err = phase.stagesStorage.PushIfNotExistSbomImage(ctx, sbomImageName); err != nil {
					return fmt.Errorf("unable to push sbom image: %q: %w", sbomImageName, err)
				}
				return nil
			} else {
				if pulled, err := phase.stagesStorage.PullIfExistSbomImage(ctx, sbomImageName); err != nil {
					return fmt.Errorf("unable to pull sbom image: %q: %w", sbomImageName, err)
				} else if pulled {
					return nil
				}
			}
		}

		// SBOM scanning is local operation. Ensure source image exist locally.
		if !phase.isLocalStorage {
			if err := phase.containerBackend.Pull(ctx, sourceImageName, container_backend.PullOpts{}); err != nil {
				return fmt.Errorf("unable to pull %q: %w", sourceImageName, err)
			}
		}

		tmpImgId, err := phase.containerBackend.GenerateSBOM(ctx, scanOpts, sbomImgLabels.ToStringSlice())
		if err != nil {
			return fmt.Errorf("unable to scan source image and store the result: %w", err)
		}

		if err = phase.containerBackend.Tag(ctx, tmpImgId, sbomImageName, container_backend.TagOpts{}); err != nil {
			return fmt.Errorf("unable to tag sbom image: %w", err)
		}

		if !phase.isLocalStorage {
			if _, err := phase.stagesStorage.PushIfNotExistSbomImage(ctx, sbomImageName); err != nil {
				return fmt.Errorf("unable to push sbom image: %q: %w", sbomImageName, err)
			}
		}

		return nil
	})
}

func (phase *sbomPhase) prepareSbomLabels(_ context.Context, srcImgLabels map[string]string, scanOpts scanner.ScanOptions) label.LabelList {
	return label.LabelList{
		label.NewLabel(image.WerfLabel, srcImgLabels[image.WerfLabel]),
		label.NewLabel(image.WerfProjectRepoCommitLabel, srcImgLabels[image.WerfProjectRepoCommitLabel]),
		label.NewLabel(image.WerfStageContentDigestLabel, srcImgLabels[image.WerfStageContentDigestLabel]),
		label.NewLabel(image.WerfSbomLabel, scanOpts.Checksum()),
		label.NewLabel(image.WerfVersionLabel, srcImgLabels[image.WerfVersionLabel]),
	}
}

func (phase *sbomPhase) findSbomImageLocally(ctx context.Context, sbomImgLabels label.LabelList, sbomImgName string) (image.Summary, bool, error) {
	sbomImgList, err := phase.containerBackend.Images(ctx, container_backend.ImagesOptions{
		Filters: filter.NewFilterListFromLabelList(sbomImgLabels[:len(sbomImgLabels)-1]).ToPairs(),
	})
	if err != nil {
		return image.Summary{}, false, fmt.Errorf("unable to list sbom images: %w", err)
	}

	_, sbomTag := image.ParseRepositoryAndTag(sbomImgName)

	img, ok := lo.Find(sbomImgList, func(img image.Summary) bool {
		return slices.ContainsFunc(img.RepoTags, func(repoTag string) bool {
			_, imgTag := image.ParseRepositoryAndTag(repoTag)
			// TODO: compare foundImgRepo and sbomImgRepo
			return imgTag == sbomTag
		})
	})

	return img, ok, nil
}
