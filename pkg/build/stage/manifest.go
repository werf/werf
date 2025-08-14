package stage

import (
	"context"
	"fmt"
	"os"

	"github.com/deckhouse/delivery-kit-sdk/pkg/integrity"
	"github.com/deckhouse/delivery-kit-sdk/pkg/signature/image"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/build/signing"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/docker_registry/api"
)

type ManifestStage struct {
	*BaseStage
	imageSpec *config.ImageSpec
}

func GenerateManifestStage(baseStageOptions *BaseStageOptions) *ManifestStage {
	return newManifestStage(baseStageOptions)
}

func newManifestStage(baseStageOptions *BaseStageOptions) *ManifestStage {
	return &ManifestStage{
		BaseStage: NewBaseStage(Manifest, baseStageOptions),
	}
}

func (s *ManifestStage) IsBuildable() bool {
	return false
}

func (s *ManifestStage) IsMutable() bool {
	return true
}

func (s *ManifestStage) PrepareImage(_ context.Context, _ Conveyor, _ container_backend.ContainerBackend, _, _ *StageImage, _ container_backend.BuildContextArchiver) error {
	return nil
}

func (s *ManifestStage) GetDependencies(_ context.Context, _ Conveyor, _ container_backend.ContainerBackend, _, _ *StageImage, _ container_backend.BuildContextArchiver) (string, error) {
	return "", nil
}

func (s *ManifestStage) MutateImage(ctx context.Context, registry docker_registry.Interface, prevBuiltImage, stageImage *StageImage, manifestSigningOptions signing.ManifestSigningOptions) error {
	srcRef := prevBuiltImage.Image.Name()
	destRef := stageImage.Image.Name()

	var opts []api.MutateOption
	if os.Getenv("WERF_DISABLE_DM_VERITY") != "1" {
		opts = append(opts, api.WithLayersMutation(func(ctx context.Context, layers []v1.Layer) ([]mutate.Addendum, error) {
			var result []mutate.Addendum
			for _, layer := range layers {
				addendum, err := integrity.AnnotateLayerWithDMVerityRootHash(ctx, layer)
				if err != nil {
					return nil, err
				}
				result = append(result, addendum)
			}

			return result, nil
		}))
	}

	if os.Getenv("WERF_DISABLE_SIGN") != "1" {
		opts = append(opts, api.WithManifestAnnotationsFunc(func(ctx context.Context, manifest *v1.Manifest) (map[string]string, error) {
			annotations, err := image.GetSignatureAnnotationsForImageManifest(ctx, manifestSigningOptions.Signer().SignerVerifier(), manifest)
			if err != nil {
				return nil, fmt.Errorf("unable to sign manifest: %w", err)
			}

			return util.MergeMaps(manifest.Annotations, annotations), nil
		}))
	}

	return registry.MutateAndPushImage(ctx, srcRef, destRef, opts...)
}
