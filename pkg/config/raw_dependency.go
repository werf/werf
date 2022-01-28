package config

type dependencyImageType string

var (
	dependencyImageTypeUnknown    dependencyImageType = ""
	dependencyImageTypeStapel     dependencyImageType = "stapel"
	dependencyImageTypeDockerfile dependencyImageType = "dockerfile"
)

type rawDependency struct {
	Image   string                 `yaml:"image,omitempty"`
	Before  string                 `yaml:"before,omitempty"`
	After   string                 `yaml:"after,omitempty"`
	Imports []*rawDependencyImport `yaml:"imports,omitempty"`

	rawStapelImage         *rawStapelImage         `yaml:"-"` // possible parent
	rawImageFromDockerfile *rawImageFromDockerfile `yaml:"-"` // possible parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (d *rawDependency) doc() *doc {
	var document *doc
	switch d.imageType() {
	case dependencyImageTypeStapel:
		document = d.rawStapelImage.doc
	case dependencyImageTypeDockerfile:
		document = d.rawImageFromDockerfile.doc
	}

	return document
}

func (d *rawDependency) UnmarshalYAML(unmarshal func(interface{}) error) error {
	switch parent := parentStack.Peek().(type) {
	case *rawStapelImage:
		d.rawStapelImage = parent
	case *rawImageFromDockerfile:
		d.rawImageFromDockerfile = parent
	}

	parentStack.Push(d)
	type plain rawDependency
	err := unmarshal((*plain)(d))
	parentStack.Pop()
	if err != nil {
		return err
	}

	if err := checkOverflow(d.UnsupportedAttributes, d, d.doc()); err != nil {
		return err
	}

	return nil
}

func (d *rawDependency) toDirective() (*Dependency, error) {
	dependency := &Dependency{
		ImageName: d.Image,
		Before:    d.Before,
		After:     d.After,
		raw:       d,
	}

	for _, rawDepImport := range d.Imports {
		depImport, err := rawDepImport.toDirective()
		if err != nil {
			return nil, err
		}

		dependency.Imports = append(dependency.Imports, depImport)
	}

	if err := dependency.validate(); err != nil {
		return nil, err
	}

	return dependency, nil
}

func (d *rawDependency) imageType() dependencyImageType {
	if d.rawStapelImage != nil {
		return dependencyImageTypeStapel
	}

	if d.rawImageFromDockerfile != nil {
		return dependencyImageTypeDockerfile
	}

	return dependencyImageTypeUnknown
}
