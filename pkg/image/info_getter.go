package image

type InfoGetter struct {
	WerfImageName string
	Tag           string
	Name          string
}

func NewInfoGetter(imageName string, name, tag string) *InfoGetter {
	return &InfoGetter{
		WerfImageName: imageName,
		Name:          name,
		Tag:           tag,
	}
}

func (d *InfoGetter) IsNameless() bool {
	return d.WerfImageName == ""
}

func (d *InfoGetter) GetWerfImageName() string {
	return d.WerfImageName
}

func (d *InfoGetter) GetName() string {
	return d.Name
}

func (d *InfoGetter) GetTag() string {
	return d.Tag
}
