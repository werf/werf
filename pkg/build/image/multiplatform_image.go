package image

import (
	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/image"
	common_image "github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/logging"
)

type MultiplatformImage struct {
	Name    string
	IsFinal bool
	Images  []*Image

	calculatedDigest string
	stageID          common_image.StageID
	stageDesc        *common_image.StageDesc
	finalStageDesc   *common_image.StageDesc

	logImageIndex int
	logImageTotal int
}

func NewMultiplatformImage(name string, images []*Image, logImageIndex, logImageTotal int) *MultiplatformImage {
	if len(images) == 0 {
		panic("expected at least one image")
	}

	img := &MultiplatformImage{
		Name:          name,
		IsFinal:       images[0].IsFinal,
		Images:        images,
		logImageIndex: logImageIndex,
		logImageTotal: logImageTotal,
	}

	metaStageDeps := util.MapFuncToSlice(images, func(img *Image) string {
		stageDesc := img.GetLastNonEmptyStage().GetStageImage().Image.GetStageDesc()
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
		stageDesc := img.GetLastNonEmptyStage().GetStageImage().Image.GetStageDesc()
		return stageDesc.Info
	})
}

func (img *MultiplatformImage) GetFinalStageDesc() *image.StageDesc {
	return img.finalStageDesc
}

func (img *MultiplatformImage) SetFinalStageDesc(stageDesc *common_image.StageDesc) {
	img.finalStageDesc = stageDesc
}

func (img *MultiplatformImage) GetStageDesc() *image.StageDesc {
	return img.stageDesc
}

func (img *MultiplatformImage) SetStageDesc(stageDesc *common_image.StageDesc) {
	img.stageDesc = stageDesc
}

func (img *MultiplatformImage) LogDetailedName() string {
	return logging.ImageLogProcessName(img.Name, img.IsFinal, "", logging.WithProgress(img.logImageIndex+1, img.logImageTotal))
}

func (img *MultiplatformImage) UseSbom() bool {
	primaryImg := img.Images[0]
	return primaryImg.UseSbom()
}
