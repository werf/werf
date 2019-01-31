package build

import (
	"fmt"
	"os"
	"strings"

	"github.com/flant/werf/pkg/build/stage"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/logger"
)

type Image struct {
	name string

	baseImageName      string
	baseImageImageName string

	stages     []stage.Interface
	baseImage  *image.StageImage
	isArtifact bool
}

func (i *Image) LogName() string {
	return imageLogName(i.name)

}

func imageLogName(name string) string {
	if name == "" {
		return "~"
	}

	return name
}

func (i *Image) SetStages(stages []stage.Interface) {
	i.stages = stages
}

func (i *Image) GetStages() []stage.Interface {
	return i.stages
}

func (i *Image) GetStage(name stage.StageName) stage.Interface {
	for _, s := range i.stages {
		if s.Name() == name {
			return s
		}
	}

	return nil
}

func (i *Image) LatestStage() stage.Interface {
	return i.stages[len(i.stages)-1]
}

func (i *Image) GetName() string {
	return i.name
}

func (i *Image) SetupBaseImage(c *Conveyor) {
	baseImageName := i.baseImageName
	if i.baseImageImageName != "" {
		baseImageName = c.GetImage(i.baseImageImageName).LatestStage().GetImage().Name()
	}

	i.baseImage = c.GetOrCreateImage(nil, baseImageName)
}

func (i *Image) GetBaseImage() *image.StageImage {
	return i.baseImage
}

func (i *Image) PrepareBaseImage(c *Conveyor) error {
	fromImage := i.stages[0].GetImage()

	if fromImage.IsExists() {
		return nil
	}

	if i.baseImageImageName != "" {
		return nil
	}

	werfImagesRegistry := os.Getenv("WERF_IMAGES_REGISTRY")
	if werfImagesRegistry != "" && strings.HasPrefix(i.baseImage.Name(), werfImagesRegistry) {
		if err := c.GetDockerAuthorizer().LoginForPull(werfImagesRegistry); err != nil {
			return fmt.Errorf("login into repo %s for base image %s failed: %s", werfImagesRegistry, i.baseImage.Name(), err)
		}
	}

	return logger.LogProcess(fmt.Sprintf("Pull %s base image", i.LogName()), "", func() error {
		if i.baseImage.IsExists() {
			if err := i.baseImage.Pull(); err != nil {
				logger.LogWarningF("WARNING: cannot pull base image %s: %s\n", i.baseImage.Name(), err)
				logger.LogWarningF("WARNING: using existing image %s without pull\n", i.baseImage.Name())
			}
			return nil
		}

		if err := i.baseImage.Pull(); err != nil {
			return fmt.Errorf("image %s pull failed: %s", i.baseImage.Name(), err)
		}

		return nil
	})
}
