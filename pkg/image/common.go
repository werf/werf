package image

import (
	"fmt"

	"github.com/flant/werf/pkg/util"
)

func GetContainerLockName(containerName string) string {
	return fmt.Sprintf("container.%s", util.Sha256Hash(containerName))
}

func GetImageLockName(imageName string) string {
	return fmt.Sprintf("image.%s", util.Sha256Hash(imageName))
}
