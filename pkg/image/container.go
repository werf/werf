package image

type Container struct {
	ID      string
	ImageID string
	Names   []string
}

func (container Container) LogName() string {
	name := container.ID
	if len(container.Names) != 0 {
		name = container.Names[0]
	}
	return name
}

type ContainerList []Container

type ContainerFilter struct {
	ID       string
	Name     string
	Ancestor string
}
