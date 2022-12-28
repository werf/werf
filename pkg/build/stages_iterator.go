package build

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
	build_image "github.com/werf/werf/pkg/build/image"
	"github.com/werf/werf/pkg/build/stage"
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
	logboek.Context(ctx).Debug().LogF("%s stage is empty: %v\n", stg.LogDetailedName(), isEmpty)

	if stg.HasPrevStage() && iterator.PrevStage == nil {
		panic(fmt.Sprintf("expected PrevStage to be set for image %q stage %s!", img.GetName(), stg.Name()))
	}

	if err := onImageStageFunc(img, stg, isEmpty); err != nil {
		return err
	}

	iterator.PrevStage = stg
	logboek.Context(ctx).Debug().LogF("Set prev stage = %q %s\n", iterator.PrevStage.Name(), iterator.PrevStage.GetDigest())

	if !isEmpty {
		iterator.PrevNonEmptyStage = stg
		logboek.Context(ctx).Debug().LogF("Set prev non empty stage = %q %s\n", iterator.PrevNonEmptyStage.Name(), iterator.PrevNonEmptyStage.GetDigest())

		if iterator.PrevNonEmptyStage.GetStageImage().Image.GetStageDescription() != nil {
			iterator.PrevNonEmptyStageImageSize = iterator.PrevNonEmptyStage.GetStageImage().Image.GetStageDescription().Info.Size
			logboek.Context(ctx).Debug().LogF("Set prev non empty stage image size = %d %q %s\n", iterator.PrevNonEmptyStageImageSize, iterator.PrevNonEmptyStage.Name(), iterator.PrevNonEmptyStage.GetDigest())
		}

		if stg.GetStageImage().Image.GetStageDescription() != nil {
			iterator.PrevBuiltStage = stg
			logboek.Context(ctx).Debug().LogF("Set prev built stage = %q (image %s)\n", iterator.PrevBuiltStage.Name(), iterator.PrevBuiltStage.GetStageImage().Image.Name())
		}
	}

	return nil
}
