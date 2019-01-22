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

func (d *Image) SetStages(stages []stage.Interface) {
	d.stages = stages
}

func (d *Image) GetStages() []stage.Interface {
	return d.stages
}

func (d *Image) GetStage(name stage.StageName) stage.Interface {
	for _, s := range d.stages {
		if s.Name() == name {
			return s
		}
	}

	return nil
}

func (d *Image) LatestStage() stage.Interface {
	return d.stages[len(d.stages)-1]
}

func (d *Image) GetName() string {
	return d.name
}

func (d *Image) SetupBaseImage(c *Conveyor) {
	baseImageName := d.baseImageName
	if d.baseImageImageName != "" {
		baseImageName = c.GetImage(d.baseImageImageName).LatestStage().GetImage().Name()
	}

	d.baseImage = c.GetOrCreateImage(nil, baseImageName)
}

func (d *Image) GetBaseImage() *image.StageImage {
	return d.baseImage
}

func (d *Image) PrepareBaseImage(c *Conveyor) error {
	fromImage := d.stages[0].GetImage()

	if fromImage.IsExists() {
		return nil
	}

	if d.baseImageImageName != "" {
		return nil
	}

	ciRegistry := os.Getenv("CI_REGISTRY")
	if ciRegistry != "" && strings.HasPrefix(d.baseImage.Name(), ciRegistry) {
		err := c.GetDockerAuthorizer().LoginForPull(ciRegistry)
		if err != nil {
			return fmt.Errorf("login into repo %s for base image %s failed: %s", ciRegistry, d.baseImage.Name(), err)
		}
	}

	if d.GetName() == "" {
		fmt.Printf("# Pulling base image for image\n")
	} else {
		fmt.Printf("# Pulling base image for image/%s\n", d.GetName())
	}

	if d.baseImage.IsExists() {
		err := d.baseImage.Pull()
		if err != nil {
			logger.LogWarningF("WARNING: cannot pull base image %s: %s\n", d.baseImage.Name(), err)
			logger.LogWarningF("WARNING: using existing image %s without pull\n", d.baseImage.Name())
		}
		return nil
	}

	err := d.baseImage.Pull()
	if err != nil {
		return fmt.Errorf("image %s pull failed: %s", d.baseImage.Name(), err)
	}

	return nil
}
