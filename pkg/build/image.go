package build

import (
	"fmt"

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
	return ImageLogName(i.name, i.isArtifact)
}

func (i *Image) LogTagName() string {
	return ImageLogTagName(i.name, i.isArtifact)
}

func ImageLogName(name string, isArtifact bool) string {
	if !isArtifact {
		if name == "" {
			name = "~"
		}
	}

	return name
}

func ImageLogTagName(name string, isArtifact bool) string {
	logName := ImageLogName(name, isArtifact)

	if isArtifact {
		return fmt.Sprintf("🔧 %s", logName)
	} else {
		return fmt.Sprintf("🚤 %s", logName)
	}
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

	return logger.LogProcess("Pulling base image", logger.LogProcessOptions{}, func() error {
		if i.baseImage.IsExists() {
			if err := i.baseImage.Pull(); err != nil {
				logger.LogErrorF("WARNING: cannot pull base image %s: %s\n", i.baseImage.Name(), err)
				logger.LogErrorF("WARNING: using existing image %s without pull\n", i.baseImage.Name())
			}
			return nil
		}

		if err := i.baseImage.Pull(); err != nil {
			return fmt.Errorf("image %s pull failed: %s", i.baseImage.Name(), err)
		}

		return nil
	})
}
