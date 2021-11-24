package build

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/container_runtime"
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

func (iterator *StagesIterator) GetPrevImage(img *Image, stg stage.Interface) container_runtime.LegacyImageInterface {
	if stg.Name() == "from" {
		return img.GetBaseImage()
	} else if iterator.PrevNonEmptyStage != nil {
		return iterator.PrevNonEmptyStage.GetImage()
	}
	return nil
}

func (iterator *StagesIterator) GetPrevBuiltImage(img *Image, stg stage.Interface) container_runtime.LegacyImageInterface {
	if stg.Name() == "from" {
		return img.GetBaseImage()
	} else if iterator.PrevBuiltStage != nil {
		return iterator.PrevBuiltStage.GetImage()
	}
	return nil
}

func (iterator *StagesIterator) OnImageStage(ctx context.Context, img *Image, stg stage.Interface, onImageStageFunc func(img *Image, stg stage.Interface, isEmpty bool) error) error {
	isEmpty, err := stg.IsEmpty(ctx, iterator.Conveyor, iterator.GetPrevBuiltImage(img, stg))
	if err != nil {
		return fmt.Errorf("error checking stage %s is empty: %s", stg.Name(), err)
	}
	logboek.Context(ctx).Debug().LogF("%s stage is empty: %v\n", stg.LogDetailedName(), isEmpty)

	if stg.Name() != "from" && stg.Name() != "dockerfile" {
		if iterator.PrevStage == nil {
			panic(fmt.Sprintf("expected PrevStage to be set for image %q stage %s!", img.GetName(), stg.Name()))
		}
	}

	if err := onImageStageFunc(img, stg, isEmpty); err != nil {
		return err
	}

	iterator.PrevStage = stg
	logboek.Context(ctx).Debug().LogF("Set prev stage = %q %s\n", iterator.PrevStage.Name(), iterator.PrevStage.GetDigest())

	if !isEmpty {
		iterator.PrevNonEmptyStage = stg
		logboek.Context(ctx).Debug().LogF("Set prev non empty stage = %q %s\n", iterator.PrevNonEmptyStage.Name(), iterator.PrevNonEmptyStage.GetDigest())

		if iterator.PrevNonEmptyStage.GetImage().GetStageDescription() != nil {
			iterator.PrevNonEmptyStageImageSize = iterator.PrevNonEmptyStage.GetImage().GetStageDescription().Info.Size
			logboek.Context(ctx).Debug().LogF("Set prev non empty stage image size = %d %q %s\n", iterator.PrevNonEmptyStageImageSize, iterator.PrevNonEmptyStage.Name(), iterator.PrevNonEmptyStage.GetDigest())
		}

		if stg.GetImage().GetStageDescription() != nil {
			iterator.PrevBuiltStage = stg
			logboek.Context(ctx).Debug().LogF("Set prev built stage = %q (image %s)\n", iterator.PrevBuiltStage.Name(), iterator.PrevBuiltStage.GetImage().Name())
		}
	}

	return nil
}
