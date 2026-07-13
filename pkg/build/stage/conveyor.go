package stage

import (
	"github.com/werf/werf/v2/pkg/giterminism_manager"
)

type Conveyor interface {
	GetImageContentTagStageID(targetPlatform, imageName string) string
	GetImageContentTagDigest(targetPlatform, imageName string) string
	GetImageContentTagName(targetPlatform, imageName string) string

	GiterminismManager() giterminism_manager.Interface
}
