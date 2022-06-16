package image

import "fmt"

type (
	CustomTagFunc func(string, string) string
	ExportTagFunc func(string, string) string
)

type InfoGetter struct {
	WerfImageName string
	Repo          string
	Tag           string

	InfoGetterOptions
}

type InfoGetterOptions struct {
	CustomTagFunc CustomTagFunc
}

func NewInfoGetter(imageName, ref string, opts InfoGetterOptions) *InfoGetter {
	repo, tag := ParseRepositoryAndTag(ref)

	return &InfoGetter{
		WerfImageName:     imageName,
		Repo:              repo,
		Tag:               tag,
		InfoGetterOptions: opts,
	}
}

func (d *InfoGetter) IsNameless() bool {
	return d.WerfImageName == ""
}

func (d *InfoGetter) GetWerfImageName() string {
	return d.WerfImageName
}

func (d *InfoGetter) GetName() string {
	return fmt.Sprintf("%s:%s", d.Repo, d.GetTag())
}

func (d *InfoGetter) GetTag() string {
	if d.CustomTagFunc != nil {
		return d.CustomTagFunc(d.WerfImageName, d.Tag)
	}
	return d.Tag
}
