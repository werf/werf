package stage

import (
	"context"
	"fmt"

	"github.com/deckhouse/delivery-kit-sdk/pkg/integrity"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"

	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/docker_registry/api"
	"github.com/werf/werf/v2/pkg/storage"
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

func (s *VerityAnnotationStage) MutateImage(ctx context.Context, stagesStorage ImageMutatorPusher, prevBuiltImage, stageImage *StageImage) error {
	// TODO: refactor type assertion by adopting stage и storage interfaces
	registry, err := registryFromImageMutatorPusher(stagesStorage)
	if err != nil {
		return err
	}

	srcRef := prevBuiltImage.Image.Name()
	destRef := stageImage.Image.Name()

	var imageDmVerityAnnotations map[string]string
	optWithLayersMutation := api.WithLayersMutation(func(ctx context.Context, layers []v1.Layer) ([]mutate.Addendum, error) {
		// annotate layers
		var result []mutate.Addendum
		for _, layer := range layers {
			annotations, err := integrity.CalculateLayerDMVerityAnnotations(ctx, layer)
			if err != nil {
				return nil, err
			}

			result = append(result, mutate.Addendum{
				Layer:       layer,
				Annotations: annotations,
			})
		}

		// save dm verity annotations for image
		{
			image, err := mutate.Append(empty.Image, result...)
			if err != nil {
				return nil, err
			}

			imageDmVerityAnnotations, err = integrity.CalculateImageDMVerityAnnotations(ctx, image)
			if err != nil {
				return nil, err
			}
		}

		return result, nil
	})

	optWithManifestAnnotationsFunc := api.WithManifestAnnotationsFunc(func(ctx context.Context, manifest *v1.Manifest) (map[string]string, error) {
		result := manifest.Annotations
		if result == nil {
			result = map[string]string{}
		}

		for key, value := range imageDmVerityAnnotations {
			result[key] = value
		}

		return result, nil
	})

	return registry.MutateAndPushImage(ctx, srcRef, destRef, optWithLayersMutation, optWithManifestAnnotationsFunc)
}

// registryFromImageMutatorPusher returns docker registry interface from stages storage.
func registryFromImageMutatorPusher(stagesStorage ImageMutatorPusher) (docker_registry.Interface, error) {
	repoStagesStorage, ok := stagesStorage.(*storage.RepoStagesStorage)
	if !ok {
		return nil, fmt.Errorf("expected stages storage of type *storage.RepoStagesStorage, got %T", stagesStorage)
	}

	return repoStagesStorage.DockerRegistry, nil
}
