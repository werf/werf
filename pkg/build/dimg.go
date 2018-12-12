package build

import (
	"fmt"
	"os"
	"strings"

	"github.com/flant/dapp/pkg/build/stage"
	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/logger"
)

type Dimg struct {
	name string

	baseImageName     string
	baseImageDimgName string

	stages     []stage.Interface
	baseImage  *image.Stage
	isArtifact bool
}

func (d *Dimg) SetStages(stages []stage.Interface) {
	d.stages = stages
}

func (d *Dimg) GetStages() []stage.Interface {
	return d.stages
}

func (d *Dimg) GetStage(name stage.StageName) stage.Interface {
	for _, s := range d.stages {
		if s.Name() == name {
			return s
		}
	}

	return nil
}

func (d *Dimg) LatestStage() stage.Interface {
	return d.stages[len(d.stages)-1]
}

func (d *Dimg) GetName() string {
	return d.name
}

func (d *Dimg) SetupBaseImage(c *Conveyor) {
	baseImageName := d.baseImageName
	if d.baseImageDimgName != "" {
		baseImageName = c.GetDimg(d.baseImageDimgName).LatestStage().GetImage().Name()
	}

	d.baseImage = c.GetOrCreateImage(nil, baseImageName)
}

func (d *Dimg) GetBaseImage() *image.Stage {
	return d.baseImage
}

func (d *Dimg) PrepareBaseImage(c *Conveyor) error {
	fromImage := d.stages[0].GetImage()

	if fromImage.IsExists() {
		return nil
	}

	if d.baseImageDimgName != "" {
		return nil
	}

	ciRegistry := os.Getenv("CI_REGISTRY")
	if ciRegistry != "" && strings.HasPrefix(fromImage.Name(), ciRegistry) {
		err := c.GetDockerAuthorizer().LoginForPull(ciRegistry)
		if err != nil {
			return fmt.Errorf("login into repo %s for base image %s failed: %s", ciRegistry, fromImage.Name(), err)
		}
	}

	if d.GetName() == "" {
		fmt.Printf("# Pulling base image for dimg\n")
	} else {
		fmt.Printf("# Pulling base image for dimg/%s\n", d.GetName())
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
