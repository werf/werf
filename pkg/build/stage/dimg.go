package stage

type Dimg struct {
	stages []Interface
}

func (d *Dimg) SetStages(stages []Interface) {
	d.stages = stages
}

func (d *Dimg) GetStages() []Interface {
	return d.stages
}

func (d *Dimg) GetStage(name string) Interface {
	for _, stage := range d.stages {
		if stage.Name() == name {
			return stage
		}
	}

	return nil
}

func (d *Dimg) LatestStage() Interface {
	return d.stages[len(d.stages)-1]
}
