package image

type Container struct {
	Name                 string
	RunCommands          []string
	ServiceRunCommands   []string
	RunOptions           *ContainerOptions
	ServiceRunOptions    *ContainerOptions
	ChangeOptions        *ContainerOptions
	ServiceChangeOptions *ContainerOptions
}

func NewContainer() *Container {
	container := &Container{}
	container.RunOptions = &ContainerOptions{}
	container.ServiceRunOptions = &ContainerOptions{}
	container.ChangeOptions = &ContainerOptions{}
	container.ServiceChangeOptions = &ContainerOptions{}
	return container
}

type ContainerOptions struct {
	Volume      []string
	VolumesFrom []string
	Expose      []string
	Env         map[string]string
	Label       map[string]string
	Cmd         []string
	Onbuild     []string
	Workdir     string
	User        string
	Entrypoint  []string
}

func (c *Container) Run() error { // TODO
	return nil
}

func (c *Container) CommitAndRm() (string, error) {
	builtId, err := c.Commit()

	if rmContainerIfExistErr := c.RmContainerIfExist(); rmContainerIfExistErr != nil {
		return "", rmContainerIfExistErr
	}

	return builtId, err
}

func (c *Container) Commit() (string, error) {
	return "", nil
}

func (c *Container) RmContainerIfExist() error {
	exist, err := c.IsExist()
	if err != nil {
		return err
	}
	if exist {
		if err := c.Rm(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Container) IsExist() (bool, error) { // TODO
	return false, nil
}

func (c *Container) Rm() error { // TODO
	return nil
}
