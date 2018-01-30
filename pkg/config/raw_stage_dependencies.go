package config

type RawStageDependencies struct {
	Install       interface{} `yaml:"install,omitempty"`
	Setup         interface{} `yaml:"setup,omitempty"`
	BeforeSetup   interface{} `yaml:"beforeSetup,omitempty"`
	BuildArtifact interface{} `yaml:"buildArtifact,omitempty"`

	RawGit *RawGit `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *RawStageDependencies) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := ParentStack.Peek().(*RawGit); ok {
		c.RawGit = parent
	}

	type plain RawStageDependencies
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := CheckOverflow(c.UnsupportedAttributes, c); err != nil {
		return err
	}

	return nil
}

func (c *RawStageDependencies) ToDirective() (stageDependencies *StageDependencies, err error) {
	stageDependencies = &StageDependencies{}

	if install, err := InterfaceToStringArray(c.Install); err != nil {
		return nil, err
	} else {
		stageDependencies.Install = install
	}

	if beforeSetup, err := InterfaceToStringArray(c.BeforeSetup); err != nil {
		return nil, err
	} else {
		stageDependencies.BeforeSetup = beforeSetup
	}

	if setup, err := InterfaceToStringArray(c.Setup); err != nil {
		return nil, err
	} else {
		stageDependencies.Setup = setup
	}

	if buildArtifact, err := InterfaceToStringArray(c.BuildArtifact); err != nil {
		return nil, err
	} else {
		stageDependencies.BuildArtifact = buildArtifact
	}

	stageDependencies.Raw = c

	if err := c.ValidateDirective(stageDependencies); err != nil {
		return nil, err
	}

	return stageDependencies, nil
}

func (c *RawStageDependencies) ValidateDirective(stageDependencies *StageDependencies) error {
	if err := stageDependencies.Validate(); err != nil {
		return err
	}

	return nil
}
