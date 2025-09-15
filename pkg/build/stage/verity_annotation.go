package stage

import (
	"context"

	"github.com/deckhouse/delivery-kit-sdk/pkg/integrity"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"

	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/docker_registry/api"
)

type VerityAnnotationStage struct {
	*BaseStage
}

func GenerateVerityAnnotationStage(baseStageOptions *BaseStageOptions) *VerityAnnotationStage {
	return newVerityAnnotationStage(baseStageOptions)
}

func newVerityAnnotationStage(baseStageOptions *BaseStageOptions) *VerityAnnotationStage {
	return &VerityAnnotationStage{
		BaseStage: NewBaseStage(VerityAnnotation, baseStageOptions),
	}
}

func (s *VerityAnnotationStage) IsBuildable() bool {
	return false
}

func (s *VerityAnnotationStage) IsMutable() bool {
	return true
}

func (s *VerityAnnotationStage) PrepareImage(_ context.Context, _ Conveyor, _ container_backend.ContainerBackend, _, _ *StageImage, _ container_backend.BuildContextArchiver) error {
	return nil
}

func (s *VerityAnnotationStage) GetDependencies(_ context.Context, _ Conveyor, _ container_backend.ContainerBackend, _, _ *StageImage, _ container_backend.BuildContextArchiver) (string, error) {
	return "", nil
}

func (s *VerityAnnotationStage) MutateImage(ctx context.Context, registry docker_registry.Interface, prevBuiltImage, stageImage *StageImage) error {
	srcRef := prevBuiltImage.Image.Name()
	destRef := stageImage.Image.Name()

	var annos map[string]string
	optWithLayersMutation := api.WithLayersMutation(func(ctx context.Context, layers []v1.Layer) ([]mutate.Addendum, error) {
		var result []mutate.Addendum
		for _, layer := range layers {
			addendum, err := integrity.AnnotateLayerWithDMVerityRootHash(ctx, layer)
			if err != nil {
				return nil, err
			}
			result = append(result, addendum)
		}

		image, err := mutate.Append(empty.Image, result...)
		if err != nil {
			return nil, err
		}

		annos, err = integrity.GetRootHashAnnotationsForImage(ctx, image)
		if err != nil {
			return nil, err
		}

		return result, nil
	})

	optWithManifestAnnotationsFunc := api.WithManifestAnnotationsFunc(func(ctx context.Context, manifest *v1.Manifest) (map[string]string, error) {
		result := manifest.Annotations
		for key, value := range annos {
			result[key] = value
		}

		return result, nil
	})

	return registry.MutateAndPushImage(ctx, srcRef, destRef, optWithLayersMutation, optWithManifestAnnotationsFunc)
}
