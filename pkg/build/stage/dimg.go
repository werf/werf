package stage

type Dimg struct{}

func (d *Dimg) GetStages() []Interface {
	return nil
}

func (d *Dimg) GetStage(name string) Interface {
	return nil
}
