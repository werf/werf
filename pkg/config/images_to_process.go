package config

import (
	"fmt"
)

type ImagesToProcess struct {
	ImageNameList []string
	WithoutImages bool
}

func NewImagesToProcess(werfConfig *WerfConfig, imageNameList []string, onlyFinal, withoutImages bool) (ImagesToProcess, error) {
	if withoutImages {
		return ImagesToProcess{ImageNameList: nil, WithoutImages: true}, nil
	}

	if len(imageNameList) > 0 {
		if err := checkImagesExistence(werfConfig, imageNameList, onlyFinal); err != nil {
			return ImagesToProcess{}, fmt.Errorf("specified images cannot be used: %w", err)
		}

		return ImagesToProcess{ImageNameList: imageNameList, WithoutImages: false}, nil
	}

	imageNameListFromConfig := werfConfig.GetImageNameList(onlyFinal)
	return ImagesToProcess{
		ImageNameList: imageNameListFromConfig,
		WithoutImages: len(imageNameListFromConfig) == 0,
	}, nil
}

func checkImagesExistence(werfConfig *WerfConfig, imageNameList []string, onlyFinal bool) error {
	for _, name := range imageNameList {
		image := werfConfig.GetImage(name)
		if image == nil {
			return fmt.Errorf("image %q not defined in werf.yaml", name)
		}

		if onlyFinal && !image.IsFinal() {
			return fmt.Errorf("image %q is not final", name)
		}
	}

	return nil
}
