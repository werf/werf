package config

import (
	"fmt"
)

type ImagesToProcess struct {
	ImageNameList      []string
	FinalImageNameList []string
	WithoutImages      bool
}

func NewImagesToProcess(werfConfig *WerfConfig, imageNameList []string, onlyFinal, withoutImages bool) (ImagesToProcess, error) {
	if withoutImages {
		return ImagesToProcess{
			ImageNameList:      nil,
			FinalImageNameList: nil,
			WithoutImages:      true,
		}, nil
	}

	if len(imageNameList) > 0 {
		var finalImageNameList []string
		for _, name := range imageNameList {
			image := werfConfig.GetImage(name)
			if image == nil {
				return ImagesToProcess{}, fmt.Errorf("specified images cannot be used: image %q not defined in werf.yaml", name)
			}

			if image.IsFinal() {
				finalImageNameList = append(finalImageNameList, name)
			} else if onlyFinal {
				return ImagesToProcess{}, fmt.Errorf("specified images cannot be used: image %q is not final", name)
			}
		}

		return ImagesToProcess{
			ImageNameList:      imageNameList,
			FinalImageNameList: finalImageNameList,
			WithoutImages:      false,
		}, nil
	}

	imageNameListFromConfig := werfConfig.GetImageNameList(onlyFinal)
	finalImageNameListFromConfig := werfConfig.GetImageNameList(true)
	return ImagesToProcess{
		ImageNameList:      imageNameListFromConfig,
		FinalImageNameList: finalImageNameListFromConfig,
		WithoutImages:      len(imageNameListFromConfig) == 0,
	}, nil
}
