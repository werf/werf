package build

import (
	"github.com/flant/dapp/pkg/build/stage"
	"github.com/flant/dapp/pkg/image"
)

type Dimg struct {
	baseImageName     string
	baseImageDimgName string

	stages    []stage.Interface
	baseImage *image.Stage
}

func (d *Dimg) SetStages(stages []stage.Interface) {
	d.stages = stages
}

func (d *Dimg) GetStages() []stage.Interface {
	return d.stages
}

func (d *Dimg) GetStage(name stage.StageName) stage.Interface {
	for _, stage := range d.stages {
		if stage.Name() == name {
			return stage
		}
	}

	return nil
}

func (d *Dimg) LatestStage() stage.Interface {
	return d.stages[len(d.stages)-1]
}

func (d *Dimg) GetName() string {
	return ""
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

func (d *Dimg) PrepareBaseImage() error {
	if d.baseImageDimgName != "" {
		return nil
	}

	return d.baseImage.Pull()
}
