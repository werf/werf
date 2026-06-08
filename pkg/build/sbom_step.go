package build

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/samber/lo"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/container_backend/filter"
	"github.com/werf/werf/v2/pkg/container_backend/label"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil"
	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil/gost"
	"github.com/werf/werf/v2/pkg/sbom/extract"
	sbomImage "github.com/werf/werf/v2/pkg/sbom/image"
	"github.com/werf/werf/v2/pkg/sbom/scanner"
	"github.com/werf/werf/v2/pkg/storage"
)

//go:generate mockgen -source sbom_step.go -package mock -destination ../../test/mock/bom_patcher.go -mock_names BOMPatcherInterface=MockBOMPatcher

type BOMPatcherInterface interface {
	Apply(ctx context.Context, bom *cdx.BOM) (*cdx.BOM, error)
}

// ErrSbomNotAvailable indicates that SBOM for the given image is not available
// (e.g. it was not built by werf, or the SBOM image is missing from the registry/local storage).
// Callers should handle this as a non-fatal condition by emitting a warning.
var ErrSbomNotAvailable = errors.New("sbom not available")

type sbomStep struct {
	containerBackend container_backend.ContainerBackend
	stagesStorage    storage.StagesStorage
	isLocalStorage   bool
}

func newSbomStep(
	backend container_backend.ContainerBackend,
	stagesStorage storage.StagesStorage,
) *sbomStep {
	_, isLocalStorage := stagesStorage.(*storage.LocalStagesStorage)

	return &sbomStep{
		containerBackend: backend,
		stagesStorage:    stagesStorage,
		isLocalStorage:   isLocalStorage,
	}
}

func (step *sbomStep) ConvergeWithMerge(ctx context.Context, werfImgName string, stageDesc *image.StageDesc, scanOpts scanner.ScanOptions, mergeOpts cyclonedxutil.MergeOpts, patchers []BOMPatcherInterface) error {
	sourceImageName := stageDesc.Info.Name
	sbomImageName := sbomImage.ImageName(sourceImageName)

	scanOpts.Commands[0].SourcePath = sourceImageName

	if err := step.prepareGostComponents(ctx, &mergeOpts); err != nil {
		return err
	}

	sbomBaseImgLabels := step.prepareSbomBaseLabelsWithMerge(ctx, stageDesc.Info.Labels, scanOpts, mergeOpts)
	sbomImgLabels := step.prepareSbomLabelsWithMerge(ctx, stageDesc.Info.Labels, scanOpts, mergeOpts)

	_, ok, err := step.findSbomImageLocally(ctx, sbomBaseImgLabels, sbomImageName)
	if err != nil {
		return err
	}

	if step.isLocalStorage {
		if ok {
			logboek.Context(ctx).Default().LogF("image %s: Use previously generated SBOM from local backend storage\n", werfImgName)

			return nil
		}
		logboek.Context(ctx).Debug().LogF("image %s: SBOM not found in local cache, will generate\n", werfImgName)
	} else {
		if ok {
			if _, err = step.stagesStorage.PushIfNotExistSbomImage(ctx, sbomImageName); err != nil {
				return fmt.Errorf("unable to push sbom image: %q: %w", sbomImageName, err)
			}

			return nil
		} else {
			if pulled, err := step.stagesStorage.PullIfExistSbomImage(ctx, sbomImageName); err != nil {
				return fmt.Errorf("unable to pull sbom image: %q: %w", sbomImageName, err)
			} else if pulled {
				logboek.Context(ctx).Default().LogF("image %s: Use previously generated SBOM from container registry\n", werfImgName)

				return nil
			}
		}
	}

	if !step.isLocalStorage {
		if err := step.containerBackend.Pull(ctx, sourceImageName, container_backend.PullOpts{}); err != nil {
			return fmt.Errorf("unable to pull %q: %w", sourceImageName, err)
		}
	}

	return logboek.Context(ctx).Default().LogProcess("image %s: SBOM processing", werfImgName).DoError(func() error {
		tmpImgId, err := step.containerBackend.GenerateSBOM(ctx, scanOpts, nil)
		if err != nil {
			return fmt.Errorf("unable to scan image: %w", err)
		}

		targetBOM, err := step.extractBOM(ctx, tmpImgId)
		if rmErr := step.containerBackend.Rmi(ctx, tmpImgId, container_backend.RmiOpts{Force: true}); rmErr != nil {
			logboek.Context(ctx).Warn().LogF("unable to remove temp image %q: %s\n", tmpImgId, rmErr)
		}
		if err != nil {
			return fmt.Errorf("unable to extract scanned BOM: %w", err)
		}

		if err := gost.Upsert(targetBOM, mergeOpts.Gost); err != nil {
			return fmt.Errorf("unable to set GOST properties into scanned BOM: %w", err)
		}

		resultBOM := targetBOM
		if !mergeOpts.IsEmpty() {
			resultBOM, err = cyclonedxutil.MergeBOMs(targetBOM, mergeOpts)
			if err != nil {
				return fmt.Errorf("merge BOMs: %w", err)
			}
		}

		for _, patcher := range patchers {
			if patcher == nil {
				continue
			}
			resultBOM, err = patcher.Apply(ctx, resultBOM)
			if err != nil {
				return err
			}
		}

		sbomImgId, err := step.buildSbomImage(ctx, resultBOM, scanOpts, sbomImgLabels.ToStringSlice())
		if err != nil {
			return err
		}

		if err = step.containerBackend.Tag(ctx, sbomImgId, sbomImageName, container_backend.TagOpts{}); err != nil {
			return fmt.Errorf("unable to tag sbom image: %w", err)
		}

		if !step.isLocalStorage {
			if _, err := step.stagesStorage.PushIfNotExistSbomImage(ctx, sbomImageName); err != nil {
				return fmt.Errorf("unable to push sbom image: %q: %w", sbomImageName, err)
			}
		}

		return nil
	})
}

// prepareGostComponents validates external SBOMs and upserts GOST properties into the user-defined fragment.
func (step *sbomStep) prepareGostComponents(ctx context.Context, mergeOpts *cyclonedxutil.MergeOpts) error {
	if !mergeOpts.Gost.AttackSurface.IsUndefined() || !mergeOpts.Gost.SecurityFunction.IsUndefined() {
		logboek.Context(ctx).Default().LogF("Warning: GOST SBOM integration is experimental and its behavior may change in the future\n")
	}

	// 1. Validate inputs early before generation.
	// Images built FROM scratch are exempt from base SBOM validation as they have no prior BOM.
	if mergeOpts.BaseBOM != nil {
		if err := gost.Validate(mergeOpts.BaseBOM); err != nil {
			return fmt.Errorf("base image SBOM validation failed: %w", err)
		}
	}
	for i, bom := range mergeOpts.ImportBOMs {
		if err := gost.Validate(bom); err != nil {
			return fmt.Errorf("imported image %d SBOM validation failed: %w", i, err)
		}
	}

	// 2. Prepare fragment early (must be done BEFORE checksum calculation).
	if mergeOpts.FragmentBOM != nil {
		if err := gost.Upsert(mergeOpts.FragmentBOM, mergeOpts.Gost); err != nil {
			return fmt.Errorf("unable to set GOST properties into user-defined fragment BOM: %w", err)
		}
	}

	return nil
}

// calculateStableChecksum generates a SHA256 identifier for labels,
// including scanning environment, merged BOM sources, and GOST configuration.
func (step *sbomStep) calculateStableChecksum(scanOpts scanner.ScanOptions, mergeOpts cyclonedxutil.MergeOpts) string {
	var parts []string
	parts = append(parts, scanOpts.Checksum())
	parts = append(parts, mergeOpts.Checksum())

	return util.Sha256Hash(strings.Join(parts, "-"))
}

func (step *sbomStep) extractBOM(ctx context.Context, imageId string) (*cdx.BOM, error) {
	opener := func() (io.ReadCloser, error) {
		return step.containerBackend.SaveImageToStream(ctx, imageId)
	}

	artifactContent, err := extract.FromImageBytes(opener)
	if err != nil {
		return nil, fmt.Errorf("unable to find SBOM artifact: %w", err)
	}

	bom, err := cyclonedxutil.BuildCycloneDX16BOMFromJSON(artifactContent)
	if err != nil {
		return nil, fmt.Errorf("unable to parse SBOM artifact: %w", err)
	}

	return bom, nil
}

func (step *sbomStep) buildSbomImage(ctx context.Context, bom *cdx.BOM, scanOpts scanner.ScanOptions, labels []string) (string, error) {
	source := container_backend.NewStaticSource(bom)
	builder := container_backend.NewSBOMImageBuilder(step.containerBackend)

	imgId, err := builder.BuildImage(ctx, source, scanOpts, labels)
	if err != nil {
		return "", fmt.Errorf("unable to build SBOM image: %w", err)
	}

	return imgId, nil
}

func (step *sbomStep) prepareSbomBaseLabelsWithMerge(_ context.Context, srcImgLabels map[string]string, scanOpts scanner.ScanOptions, mergeOpts cyclonedxutil.MergeOpts) label.LabelList {
	checksum := step.calculateStableChecksum(scanOpts, mergeOpts)

	return label.LabelList{
		label.NewLabel(image.WerfLabel, srcImgLabels[image.WerfLabel]),
		label.NewLabel(image.WerfProjectRepoCommitLabel, srcImgLabels[image.WerfProjectRepoCommitLabel]),
		label.NewLabel(image.WerfStageContentDigestLabel, srcImgLabels[image.WerfStageContentDigestLabel]),
		label.NewLabel(image.WerfSbomLabel, checksum),
	}
}

func (step *sbomStep) prepareSbomLabelsWithMerge(ctx context.Context, srcImgLabels map[string]string, scanOpts scanner.ScanOptions, mergeOpts cyclonedxutil.MergeOpts) label.LabelList {
	list := step.prepareSbomBaseLabelsWithMerge(ctx, srcImgLabels, scanOpts, mergeOpts)
	list.Add(label.NewLabel(image.WerfVersionLabel, srcImgLabels[image.WerfVersionLabel]))

	return list
}

func (step *sbomStep) findSbomImageLocally(ctx context.Context, sbomBaseImgLabels label.LabelList, sbomImgName string) (image.Summary, bool, error) {
	sbomImgList, err := step.containerBackend.Images(ctx, container_backend.ImagesOptions{
		Filters: filter.NewFilterListFromLabelList(sbomBaseImgLabels).ToPairs(),
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

func (step *sbomStep) GetImageBOM(ctx context.Context, werfImgName, imageRef string, imageInfo *image.Info) (*cdx.BOM, error) {
	if imageInfo == nil {
		return nil, fmt.Errorf("image info not available for %q: %w", imageRef, ErrSbomNotAvailable)
	}

	return step.pullImageSbom(ctx, werfImgName, imageInfo)
}

func (step *sbomStep) pullImageSbom(ctx context.Context, werfImgName string, imageInfo *image.Info) (*cdx.BOM, error) {
	sbomImageName, err := step.resolveImageSbomName(imageInfo)
	if err != nil {
		return nil, err
	}

	if err = logboek.Context(ctx).Default().LogProcess("image %s: image SBOM processing (%s)", werfImgName, imageInfo.Name).DoError(func() error {
		return step.ensureSbomImageExists(ctx, sbomImageName, imageInfo.Name)
	}); err != nil {
		return nil, fmt.Errorf("unable to pull image SBOM: %w", err)
	}

	opener := func() (io.ReadCloser, error) {
		return step.containerBackend.SaveImageToStream(ctx, sbomImageName)
	}

	artifactContent, err := extract.FromImageBytes(opener)
	if err != nil {
		return nil, fmt.Errorf("unable to find SBOM artifact: %w", err)
	}

	bom, err := cyclonedxutil.BuildCycloneDX16BOMFromJSON(artifactContent)
	if err != nil {
		return nil, fmt.Errorf("unable to parse SBOM artifact: %w", err)
	}

	return bom, nil
}

func (step *sbomStep) resolveImageSbomName(baseImageInfo *image.Info) (string, error) {
	if digest, ok := baseImageInfo.Labels[image.WerfStageContentDigestLabel]; ok && digest != "" {
		_, tag := image.ParseRepositoryAndTag(baseImageInfo.Name)

		return sbomImage.BaseImageName(baseImageInfo.Repository, tag), nil
	}

	return "", fmt.Errorf(
		"unable to resolve SBOM name for image %q: required werf stage content digest label is missing: %w",
		baseImageInfo.Name, ErrSbomNotAvailable,
	)
}

func (step *sbomStep) ensureSbomImageExists(ctx context.Context, sbomImageName, sourceImageName string) error {
	if info, err := step.containerBackend.GetImageInfo(ctx, sbomImageName, container_backend.GetImageInfoOpts{}); err == nil && info != nil {
		logboek.Context(ctx).Default().LogF("Using local image SBOM %s\n", sbomImageName)

		return nil
	}

	if step.isLocalStorage {
		return fmt.Errorf("SBOM for image %q not found locally: %w", sourceImageName, ErrSbomNotAvailable)
	}

	logboek.Context(ctx).Default().LogF("Pulling image SBOM from %s\n", sbomImageName)
	if err := step.containerBackend.Pull(ctx, sbomImageName, container_backend.PullOpts{}); err != nil {
		return fmt.Errorf("SBOM for image %q not found in container registry (%w): %w", sourceImageName, err, ErrSbomNotAvailable)
	}

	return nil
}
