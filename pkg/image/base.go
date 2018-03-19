package image

import "github.com/docker/docker/api/types"

type Base struct {
	Name    string
	Id      string
	Inspect types.ImageInspect
}

func NewBaseImage() *Base {
	return &Base{}
}

func (i *Base) GetId() string {
	if i.Id != "" {
		return i.Id
	} else {
		return "" // TODO
	}
}

func (i *Base) IsTagged() bool {
	return false
}

func (i *Base) Untag() error {
	return nil
}

func (i *Base) Push() error {
	return nil
}

func (i *Base) Pull() error {
	return nil
}
