package common

import (
	"github.com/werf/werf/pkg/build"
	"github.com/werf/werf/pkg/config"
)

func GetImagesToProcess(onlyImages []string, withoutImages bool) build.ImagesToProcess {
	if withoutImages {
		return build.NewImagesToProcess(nil, true)
	} else if len(onlyImages) > 0 {
		return build.NewImagesToProcess(onlyImages, false)
	}
	return build.NewImagesToProcess(nil, false)
}

func GetImageNameList(imagesToProcess build.ImagesToProcess, werfConfig *config.WerfConfig) []string {
	if imagesToProcess.WithoutImages {
		return []string{}
	}

	if len(imagesToProcess.OnlyImages) != 0 {
		return imagesToProcess.OnlyImages
	}

	return werfConfig.GetImageNameList()
}
