package common

import "github.com/werf/werf/pkg/build"

func GetImagesToProcess(onlyImages []string, withoutImages bool) build.ImagesToProcess {
	if withoutImages {
		return build.NewImagesToProcess(nil, true)
	} else if len(onlyImages) > 0 {
		return build.NewImagesToProcess(onlyImages, false)
	}
	return build.NewImagesToProcess(nil, false)
}
