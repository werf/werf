package image

import (
	"github.com/werf/werf/pkg/image"
	common_image "github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/storage/manager"
	"github.com/werf/werf/pkg/util"
)

type MultiplatformImage struct {
	Name   string
	Images []*Image

	MultiplatformImageOptions

	calculatedDigest      string
	stageID               common_image.StageID
	stageDescription      *common_image.StageDescription
	finalStageDescription *common_image.StageDescription
}

type MultiplatformImageOptions struct {
	IsArtifact, IsDockerfileImage, IsDockerfileTargetStage bool
}

func NewMultiplatformImage(name string, images []*Image, storageManager manager.StorageManagerInterface, opts MultiplatformImageOptions) *MultiplatformImage {
	img := &MultiplatformImage{
		Name:                      name,
		Images:                    images,
		MultiplatformImageOptions: opts,
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

func (img *MultiplatformImage) IsFinal() bool {
	if img.IsArtifact {
		return false
	}
	if img.IsDockerfileImage && !img.IsDockerfileTargetStage {
		return false
	}
	return true
}
