package image

type Stage struct {
	*Base
	From      *Stage
	BuiltId   string
	Container *Container
}

func NewStageImage(from *Stage, name string, builtId string) *Stage {
	stage := &Stage{}
	stage.Name = name
	stage.From = from
	stage.BuiltId = builtId
	stage.Base = NewBaseImage()
	stage.Container = NewContainer()
	return stage
}

func (i *Stage) NewStageImage(name string) *Stage {
	stage := NewStageImage(i.From, i.Name, i.BuiltId)
	return stage
}

func (i *Stage) IsBuilt() bool {
	if i.GetBuiltId() != "" {
		return true
	}
	return false
}

func (i *Stage) GetBuiltId() string {
	if i.BuiltId != "" {
		return i.BuiltId
	} else {
		return i.GetId()
	}
}

func (i *Stage) Build() error {
	if err := i.Container.Run(); err != nil {
		return err
	}

	if builtId, err := i.Container.CommitAndRm(); err != nil {
		return err
	} else {
		i.BuiltId = builtId
	}

	return nil
}

func (i *Stage) SaveInCache() error {
	return Tag(i.BuiltId, i.Name)
}

func (i *Stage) Tag(tag string) error {
	return Tag(i.BuiltId, tag)
}

func (i *Stage) Import(name string) error {
	stage := i.NewStageImage(name)

	if err := stage.Pull(); err != nil {
		return err
	}

	i.BuiltId = stage.BuiltId

	if err := i.SaveInCache(); err != nil {
		return err
	}

	if err := stage.Untag(); err != nil {
		return err
	}

	return nil
}

func (i *Stage) Export(name string) error {
	stage := i.NewStageImage(name)

	if err := stage.Push(); err != nil {
		return err
	}

	if err := stage.Untag(); err != nil {
		return err
	}

	return nil
}

func Tag(builtId string, tag string) error { // TODO
	return nil
}
