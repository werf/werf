package image

import (
	"github.com/werf/werf/v2/pkg/image"
	common_image "github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/util"
)

type MultiplatformImage struct {
	Name    string
	IsFinal bool
	Images  []*Image

	calculatedDigest      string
	stageID               common_image.StageID
	stageDescription      *common_image.StageDescription
	finalStageDescription *common_image.StageDescription
}

func NewMultiplatformImage(name string, images []*Image) *MultiplatformImage {
	if len(images) == 0 {
		panic("expected at least one image")
	}

	img := &MultiplatformImage{
		Name:    name,
		IsFinal: images[0].IsFinal,
		Images:  images,
	}

	metaStageDeps := util.MapFuncToSlice(images, func(img *Image) string {
		stageDesc := img.GetLastNonEmptyStage().GetStageImage().Image.GetStageDescription()
		return stageDesc.StageID.String()
	})
	img.calculatedDigest = util.Sha3_224Hash(metaStageDeps...)
	img.stageID = *common_image.NewStageID(img.GetDigest(), 0)

	return img
}

func (img *MultiplatformImage) GetPlatforms() []string {
	return util.MapFuncToSlice(img.Images, func(img *Image) string { return img.TargetPlatform })
}

func (img *MultiplatformImage) GetDigest() string {
	return img.calculatedDigest
}

func (img *MultiplatformImage) GetStageID() common_image.StageID {
	return img.stageID
}

func (img *MultiplatformImage) GetImagesInfoList() []*common_image.Info {
	return util.MapFuncToSlice(img.Images, func(img *Image) *common_image.Info {
		stageDesc := img.GetLastNonEmptyStage().GetStageImage().Image.GetStageDescription()
		return stageDesc.Info
	})
}

func (img *MultiplatformImage) GetFinalStageDescription() *image.StageDescription {
	return img.finalStageDescription
}

func (img *MultiplatformImage) SetFinalStageDescription(desc *common_image.StageDescription) {
	img.finalStageDescription = desc
}

func (img *MultiplatformImage) GetStageDescription() *image.StageDescription {
	return img.stageDescription
}

func (img *MultiplatformImage) SetStageDescription(desc *common_image.StageDescription) {
	img.stageDescription = desc
}
