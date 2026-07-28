package build

import (
	"context"
	"fmt"

	build_image "github.com/werf/werf/v2/pkg/build/image"
	"github.com/werf/werf/v2/pkg/build/stage"
)

type StagesIterator struct {
	Conveyor *Conveyor

	PrevStage                  stage.Interface
	PrevNonEmptyStage          stage.Interface
	PrevBuiltStage             stage.Interface
	PrevNonEmptyStageImageSize int64
}

func NewStagesIterator(conveyor *Conveyor) *StagesIterator {
	return &StagesIterator{Conveyor: conveyor}
}

func (iterator *StagesIterator) GetPrevImage(img *build_image.Image, stg stage.Interface) *stage.StageImage {
	if stg.HasPrevStage() {
		return iterator.PrevNonEmptyStage.GetStageImage()
	} else if stg.IsStapelStage() && stg.Name() == "from" {
		return img.GetBaseStageImage()
	} else if img.IsDockerfileImage && img.DockerfileImageConfig.Staged {
		return img.GetBaseStageImage()
	}
	return nil
}

func (iterator *StagesIterator) GetPrevBuiltImage(img *build_image.Image, stg stage.Interface) *stage.StageImage {
	if stg.HasPrevStage() {
		return iterator.PrevBuiltStage.GetStageImage()
	} else if stg.IsStapelStage() && stg.Name() == "from" {
		return img.GetBaseStageImage()
	} else if img.IsDockerfileImage && img.DockerfileImageConfig.Staged {
		return img.GetBaseStageImage()
	}
	return nil
}

func (iterator *StagesIterator) OnImageStage(ctx context.Context, img *build_image.Image, stg stage.Interface, onImageStageFunc func(img *build_image.Image, stg stage.Interface, isEmpty bool) error) error {
	isEmpty, err := stg.IsEmpty(ctx, iterator.Conveyor, iterator.GetPrevBuiltImage(img, stg)) // FIXME(stapel-to-buildah): use StageImage
	if err != nil {
		return fmt.Errorf("error checking stage %s is empty: %w", stg.Name(), err)
	}

	if stg.HasPrevStage() && iterator.PrevStage == nil {
		panic(fmt.Sprintf("expected PrevStage to be set for image %q stage %s!", img.GetName(), stg.Name()))
	}

	if err := onImageStageFunc(img, stg, isEmpty); err != nil {
		return err
	}

	iterator.PrevStage = stg

	if !isEmpty {
		iterator.PrevNonEmptyStage = stg

		if iterator.PrevNonEmptyStage.GetStageImage().Image.GetStageDesc() != nil {
			iterator.PrevNonEmptyStageImageSize = iterator.PrevNonEmptyStage.GetStageImage().Image.GetStageDesc().Info.Size
		}

		if stg.GetStageImage().Image.GetStageDesc() != nil {
			iterator.PrevBuiltStage = stg
		}
	}

	return nil
}
