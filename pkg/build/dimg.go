package build

import (
	"github.com/flant/dapp/pkg/build/stage"
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/image"
)

type Dimg struct {
	baseImageName     string
	baseImageDimgName string

	stages []stage.Interface
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

func (d *Dimg) GetConfig() *config.Dimg {
	return nil
}

func (d *Dimg) CreateBaseImage(c *Conveyor) *image.Stage {
	baseImageName := d.baseImageName
	if d.baseImageDimgName != "" {
		baseImageName = c.GetDimg(d.baseImageDimgName).LatestStage().GetImage().GetName()
	}

	return c.GetOrCreateImage(nil, baseImageName)
}
