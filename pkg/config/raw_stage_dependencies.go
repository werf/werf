package config

type rawStageDependencies struct {
	Install     interface{} `yaml:"install,omitempty"`
	Setup       interface{} `yaml:"setup,omitempty"`
	BeforeSetup interface{} `yaml:"beforeSetup,omitempty"`

	rawGit *rawGit `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *rawStageDependencies) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := parentStack.Peek().(*rawGit); ok {
		c.rawGit = parent
	}

	type plain rawStageDependencies
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, c, c.rawGit.rawStapelImage.doc); err != nil {
		return err
	}

	return nil
}

func (c *rawStageDependencies) toDirective() (stageDependencies *StageDependencies, err error) {
	stageDependencies = &StageDependencies{}

	if install, err := InterfaceToStringArray(c.Install, c, c.rawGit.rawStapelImage.doc); err != nil {
		return nil, err
	} else {
		stageDependencies.Install = install
	}

	if beforeSetup, err := InterfaceToStringArray(c.BeforeSetup, c, c.rawGit.rawStapelImage.doc); err != nil {
		return nil, err
	} else {
		stageDependencies.BeforeSetup = beforeSetup
	}

	if setup, err := InterfaceToStringArray(c.Setup, c, c.rawGit.rawStapelImage.doc); err != nil {
		return nil, err
	} else {
		stageDependencies.Setup = setup
	}

	stageDependencies.raw = c

	if err := c.validateDirective(stageDependencies); err != nil {
		return nil, err
	}

	return stageDependencies, nil
}

func (c *rawStageDependencies) validateDirective(stageDependencies *StageDependencies) error {
	if err := stageDependencies.validate(); err != nil {
		return err
	}

	return nil
}
