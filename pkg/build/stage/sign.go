package stage

import (
	"context"
	"fmt"

	"github.com/deckhouse/delivery-kit-sdk/pkg/signature/image"
	v1 "github.com/google/go-containerregistry/pkg/v1"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/build/signing"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/docker_registry/api"
)

type SignStage struct {
	*BaseStage

	manifestSigningOptions signing.ManifestSigningOptions
}

func GenerateSignStage(baseStageOptions *BaseStageOptions, manifestSigningOptions signing.ManifestSigningOptions) *SignStage {
	return newSignStage(baseStageOptions, manifestSigningOptions)
}

func newSignStage(baseStageOptions *BaseStageOptions, manifestSigningOptions signing.ManifestSigningOptions) *SignStage {
	return &SignStage{
		BaseStage:              NewBaseStage(Sign, baseStageOptions),
		manifestSigningOptions: manifestSigningOptions,
	}
}

func (s *SignStage) IsBuildable() bool {
	return false
}

func (s *SignStage) IsMutable() bool {
	return true
}

func (s *SignStage) PrepareImage(_ context.Context, _ Conveyor, _ container_backend.ContainerBackend, _, _ *StageImage, _ container_backend.BuildContextArchiver) error {
	return nil
}

func (s *SignStage) GetDependencies(_ context.Context, _ Conveyor, _ container_backend.ContainerBackend, _, _ *StageImage, _ container_backend.BuildContextArchiver) (string, error) {
	args := []string{
		s.manifestSigningOptions.Signer().Cert(),
		s.manifestSigningOptions.Signer().Chain(),
	}

	return util.Sha256Hash(args...), nil
}

func (s *SignStage) MutateImage(ctx context.Context, registry docker_registry.Interface, prevBuiltImage, stageImage *StageImage) error {
	srcRef := prevBuiltImage.Image.Name()
	destRef := stageImage.Image.Name()

	opt := api.WithManifestAnnotationsFunc(func(ctx context.Context, manifest *v1.Manifest) (map[string]string, error) {
		annotations, err := image.GetSignatureAnnotationsForImageManifest(ctx, s.manifestSigningOptions.Signer().SignerVerifier(), manifest)
		if err != nil {
			return nil, fmt.Errorf("unable to sign manifest: %w", err)
		}

		return util.MergeMaps(manifest.Annotations, annotations), nil
	})

	return registry.MutateAndPushImage(ctx, srcRef, destRef, opt)
}
