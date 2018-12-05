package stage

import (
	"github.com/flant/dapp/pkg/build/builder"
	"github.com/flant/dapp/pkg/image"
)

func newUserWithGAPatchStage(builder builder.Builder, baseStageOptions *NewBaseStageOptions) *UserWithGAPatchStage {
	s := &UserWithGAPatchStage{}
	s.UserStage = newUserStage(builder, baseStageOptions)
	s.GAPatchStage = newGAPatchStage(baseStageOptions)
	return s
}

type UserWithGAPatchStage struct {
	*UserStage
	GAPatchStage *GAPatchStage
}

func (s *UserWithGAPatchStage) PrepareImage(c Conveyor, prevBuiltImage, image image.Image) error {
	if err := s.BaseStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	if err := s.GAPatchStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return nil
	}

	return nil
}
