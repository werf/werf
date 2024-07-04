package common

import (
	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/config"
)

func GetImagesToProcess(imageNameList []string, withoutImages bool) build.ImagesToProcess {
	if withoutImages {
		return build.NewImagesToProcess(nil, true)
	} else if len(imageNameList) > 0 {
		return build.NewImagesToProcess(imageNameList, false)
	}
	return build.NewImagesToProcess(nil, false)
}

func GetImageNameList(imagesToProcess build.ImagesToProcess, werfConfig *config.WerfConfig) []string {
	if imagesToProcess.WithoutImages {
		return []string{}
	}

	if len(imagesToProcess.ImageNameList) > 0 {
		return imagesToProcess.ImageNameList
	}

	return werfConfig.GetImageNameList(true)
}
