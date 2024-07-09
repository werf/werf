package build

import (
	"fmt"

	"github.com/werf/werf/pkg/config"
)

type ImagesToProcess struct {
	ImageNameList []string
	WithoutImages bool
}

func (i *ImagesToProcess) CheckImagesExistence(werfConfig *config.WerfConfig) error {
	if err := werfConfig.CheckImagesExistence(i.ImageNameList, true); err != nil {
		return fmt.Errorf("specified images cannot be used: %w", err)
	}
	return nil
}

func (i *ImagesToProcess) HaveImagesToProcess(werfConfig *config.WerfConfig) bool {
	return !i.WithoutImages && len(werfConfig.Images(true)) > 0
}

func NewImagesToProcess(imageNameList []string, withoutImages bool) ImagesToProcess {
	return ImagesToProcess{ImageNameList: imageNameList, WithoutImages: withoutImages}
}
