package stage

import (
	"context"
	"fmt"
	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/docker_registry/api"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"

	"github.com/deckhouse/delivery-kit-sdk/pkg/dmverity"
	"github.com/deckhouse/delivery-kit-sdk/pkg/signature/image"
	"github.com/deckhouse/delivery-kit-sdk/pkg/signver"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker_registry"
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

func (s *ManifestStage) MutateImage(ctx context.Context, registry docker_registry.Interface, prevBuiltImage, stageImage *StageImage) error {
	srcRef := prevBuiltImage.Image.Name()
	destRef := stageImage.Image.Name()

	return registry.MutateAndPushImage(
		ctx,
		srcRef,
		destRef,
		api.WithLayersMutation(func(ctx context.Context, layers []v1.Layer) ([]mutate.Addendum, error) {
			var result []mutate.Addendum
			for _, layer := range layers {
				addendum, err := dmverity.AnnotateLayerWithDMVerityRootHash(ctx, layer)
				if err != nil {
					return nil, err
				}
				result = append(result, addendum)
			}

			return result, nil
		}),
		api.WithManifestAnnotationsFunc(func(ctx context.Context, manifest *v1.Manifest) (map[string]string, error) {
			sv, err := signver.NewSignerVerifier(
				ctx,
				"cosign.crt",
				"cosign-chain.pem",
				signver.KeyOpts{
					KeyRef: "cosign2.key.key",
				},
			)
			if err != nil {
				return nil, fmt.Errorf("unable to load signer verifier: %w", err)
			}

			annotations, err := image.GetSignatureAnnotationsForImageManifest(ctx, sv, manifest)
			if err != nil {
				return nil, fmt.Errorf("unable to sign manifest: %w", err)
			}

			return util.MergeMaps(manifest.Annotations, annotations), nil
		}),
	)
}
