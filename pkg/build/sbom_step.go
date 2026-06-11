package build

import (
	"context"
	"errors"
	"fmt"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/oci/artifact"
	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil"
	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil/gost"
	sbomImage "github.com/werf/werf/v2/pkg/sbom/image"
	"github.com/werf/werf/v2/pkg/sbom/scanner"
	"github.com/werf/werf/v2/pkg/storage"
)

//go:generate mockgen -source sbom_step.go -package mock -destination ../../test/mock/bom_patcher.go -mock_names BOMPatcherInterface=MockBOMPatcher

type BOMPatcherInterface interface {
	Apply(ctx context.Context, bom *cdx.BOM) (*cdx.BOM, error)
}

// ErrSbomNotRequired indicates that SBOM is intentionally absent for the image
// (e.g. it is a trusted builder image). Callers should handle this silently.
var ErrSbomNotRequired = errors.New("sbom not required")

type sbomStep struct {
	containerBackend container_backend.ContainerBackend
	stagesStorage    storage.StagesStorage
}

func newSbomStep(
	backend container_backend.ContainerBackend,
	stagesStorage storage.StagesStorage,
) *sbomStep {
	return &sbomStep{
		containerBackend: backend,
		stagesStorage:    stagesStorage,
	}
}

func (step *sbomStep) ConvergeWithMerge(ctx context.Context, werfImgName string, stageDesc *image.StageDesc, scanOpts scanner.ScanOptions, mergeOpts cyclonedxutil.MergeOpts, patchers []BOMPatcherInterface, targetPlatform string) error {
	repo := stageDesc.Info.Repository
	parentDigest := stageDesc.Info.GetDigest()

	scanOpts.Commands[0].SourcePath = stageDesc.Info.Name

	if err := step.prepareGostComponents(ctx, &mergeOpts); err != nil {
		return err
	}

	checksum := step.calculateStableChecksum(scanOpts, mergeOpts)

	store := artifact.NewOCIStore(repo, werfImgName)
	desc, found, err := store.GetAttached(ctx, parentDigest, sbomImage.DSSEMediaType)
	if err != nil {
		return fmt.Errorf("check SBOM cache: %w", err)
	}
	if found && desc.Annotations[image.WerfChecksumAnnotation] == checksum {
		logboek.Context(ctx).Default().LogF("image %s: Use previously generated SBOM from registry\n", werfImgName)
		return nil
	}

	if err := step.containerBackend.Pull(ctx, stageDesc.Info.Name, container_backend.PullOpts{}); err != nil {
		return fmt.Errorf("unable to pull %q: %w", stageDesc.Info.Name, err)
	}

	return logboek.Context(ctx).Default().LogProcess("image %s: SBOM processing", werfImgName).DoError(func() error {
		bomJSON, err := step.containerBackend.GenerateSBOM(ctx, scanOpts)
		if err != nil {
			return fmt.Errorf("generate SBOM: %w", err)
		}

		targetBOM, err := cyclonedxutil.BuildCycloneDX16BOMFromJSON(bomJSON)
		if err != nil {
			return fmt.Errorf("parse scanned BOM: %w", err)
		}

		if err := gost.Upsert(targetBOM, mergeOpts.Gost); err != nil {
			return fmt.Errorf("set GOST properties: %w", err)
		}

		resultBOM := targetBOM
		if !mergeOpts.IsEmpty() {
			var err error
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

		resultJSON, err := cyclonedxutil.ToJSON(resultBOM)
		if err != nil {
			return fmt.Errorf("serialize BOM: %w", err)
		}

		if err := logboek.Context(ctx).Default().LogProcess("Push SBOM artifact").DoError(func() error {
			return sbomImage.PushSBOM(ctx, resultJSON, repo, parentDigest, werfImgName, checksum, targetPlatform)
		}); err != nil {
			return err
		}

		return nil
	})
}

func (step *sbomStep) calculateStableChecksum(scanOpts scanner.ScanOptions, mergeOpts cyclonedxutil.MergeOpts) string {
	var parts []string
	parts = append(parts, scanOpts.Checksum())
	parts = append(parts, mergeOpts.Checksum())
	return util.Sha256Hash(strings.Join(parts, "-"))
}

func (step *sbomStep) GetImageBOM(ctx context.Context, imageName string, imageInfo *image.Info) (*cdx.BOM, error) {
	if imageInfo == nil {
		return nil, fmt.Errorf("image info is nil for %q", imageName)
	}

	bom, err := step.pullImageSbom(ctx, imageName, imageInfo)
	if err != nil {
		if !errors.Is(err, artifact.ErrNotFound) {
			return nil, fmt.Errorf("pull SBOM for %q: %w", imageName, err)
		}
		if isTrustedBuilderImage(imageInfo.Labels) {
			return nil, fmt.Errorf("trusted builder image %q: %w", imageName, ErrSbomNotRequired)
		}
		return nil, fmt.Errorf("pull SBOM for %q: %w", imageName, err)
	}

	return bom, nil
}

func (step *sbomStep) pullImageSbom(ctx context.Context, imageName string, imageInfo *image.Info) (*cdx.BOM, error) {
	parentDigest := imageInfo.GetDigest()
	if parentDigest == "" {
		return nil, fmt.Errorf("image digest not available for %q", imageInfo.Name)
	}

	bomJSON, err := sbomImage.PullSBOM(ctx, imageInfo.Repository, parentDigest, imageName)
	if err != nil {
		return nil, fmt.Errorf("pull SBOM for %q: %w", imageName, err)
	}

	bom, err := cyclonedxutil.BuildCycloneDX16BOMFromJSON(bomJSON)
	if err != nil {
		return nil, fmt.Errorf("parse CycloneDX BOM: %w", err)
	}

	return bom, nil
}

func isTrustedBuilderImage(labels map[string]string) bool {
	if labels == nil {
		return false
	}
	return labels[image.DeckhouseInternalBuilderLabel] == "true"
}

func (step *sbomStep) prepareGostComponents(ctx context.Context, mergeOpts *cyclonedxutil.MergeOpts) error {
	if !mergeOpts.Gost.AttackSurface.IsUndefined() || !mergeOpts.Gost.SecurityFunction.IsUndefined() {
		logboek.Context(ctx).Default().LogF("Warning: GOST SBOM integration is experimental and its behavior may change in the future\n")
	}

	if mergeOpts.BaseBOM != nil {
		if err := gost.Validate(mergeOpts.BaseBOM); err != nil {
			return fmt.Errorf("base SBOM validation failed: %w", err)
		}
		if err := gost.Upsert(mergeOpts.BaseBOM, mergeOpts.Gost); err != nil {
			return fmt.Errorf("set GOST properties for base SBOM: %w", err)
		}
	}

	for i, externalBOM := range mergeOpts.ImportBOMs {
		if err := gost.Validate(externalBOM); err != nil {
			return fmt.Errorf("external SBOM [%d] validation failed: %w", i, err)
		}
		if err := gost.Upsert(externalBOM, mergeOpts.Gost); err != nil {
			return fmt.Errorf("set GOST properties for external SBOM [%d]: %w", i, err)
		}
	}

	if mergeOpts.FragmentBOM != nil {
		if err := gost.Upsert(mergeOpts.FragmentBOM, mergeOpts.Gost); err != nil {
			return fmt.Errorf("set GOST properties for fragment BOM: %w", err)
		}
	}

	return nil
}
