package config

type rawDependencyImport struct {
	Type           string `yaml:"type,omitempty"`
	TargetBuildArg string `yaml:"targetBuildArg,omitempty"`
	TargetEnv      string `yaml:"targetEnv,omitempty"`

	rawDependency *rawDependency `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (i *rawDependencyImport) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := parentStack.Peek().(*rawDependency); ok {
		i.rawDependency = parent
	}

	type plain rawDependencyImport
	if err := unmarshal((*plain)(i)); err != nil {
		return err
	}

	if err := checkOverflow(i.UnsupportedAttributes, i, i.rawDependency.doc()); err != nil {
		return err
	}

	return nil
}

func (i *rawDependencyImport) toDirective() (*DependencyImport, error) {
	depImport := &DependencyImport{
		Type:           DependencyImportType(i.Type),
		TargetBuildArg: i.TargetBuildArg,
		TargetEnv:      i.TargetEnv,
		raw:            i,
	}

	if err := depImport.validate(); err != nil {
		return nil, err
	}

	return depImport, nil
}
