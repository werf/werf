package config

type rawAnsible struct {
	BeforeInstall             []rawAnsibleTask `yaml:"beforeInstall"`
	Install                   []rawAnsibleTask `yaml:"install"`
	BeforeSetup               []rawAnsibleTask `yaml:"beforeSetup"`
	Setup                     []rawAnsibleTask `yaml:"setup"`
	CacheVersion              string           `yaml:"cacheVersion,omitempty"`
	BeforeInstallCacheVersion string           `yaml:"beforeInstallCacheVersion,omitempty"`
	InstallCacheVersion       string           `yaml:"installCacheVersion,omitempty"`
	BeforeSetupCacheVersion   string           `yaml:"beforeSetupCacheVersion,omitempty"`
	SetupCacheVersion         string           `yaml:"setupCacheVersion,omitempty"`

	rawImage *rawStapelImage `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *rawAnsible) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := parentStack.Peek().(*rawStapelImage); ok {
		c.rawImage = parent
	}

	parentStack.Push(c)
	type plain rawAnsible
	err := unmarshal((*plain)(c))
	parentStack.Pop()
	if err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, c, c.rawImage.doc); err != nil {
		return err
	}

	return nil
}

func (c *rawAnsible) toDirective() (ansible *Ansible, err error) {
	ansible = &Ansible{}

	ansible.CacheVersion = c.CacheVersion
	ansible.BeforeInstallCacheVersion = c.BeforeInstallCacheVersion
	ansible.InstallCacheVersion = c.InstallCacheVersion
	ansible.BeforeSetupCacheVersion = c.BeforeSetupCacheVersion
	ansible.SetupCacheVersion = c.SetupCacheVersion

	for ind := range c.BeforeInstall {
		if ansibleTask, err := c.BeforeInstall[ind].toDirective(); err != nil {
			return nil, err
		} else {
			ansible.BeforeInstall = append(ansible.BeforeInstall, ansibleTask)
		}
	}

	for ind := range c.Install {
		if ansibleTask, err := c.Install[ind].toDirective(); err != nil {
			return nil, err
		} else {
			ansible.Install = append(ansible.Install, ansibleTask)
		}
	}

	for ind := range c.BeforeSetup {
		if ansibleTask, err := c.BeforeSetup[ind].toDirective(); err != nil {
			return nil, err
		} else {
			ansible.BeforeSetup = append(ansible.BeforeSetup, ansibleTask)
		}
	}

	for ind := range c.Setup {
		if ansibleTask, err := c.Setup[ind].toDirective(); err != nil {
			return nil, err
		} else {
			ansible.Setup = append(ansible.Setup, ansibleTask)
		}
	}

	ansible.raw = c

	if err := c.validateDirective(ansible); err != nil {
		return nil, err
	}

	return ansible, nil
}

func (c *rawAnsible) validateDirective(ansible *Ansible) (err error) {
	if err := ansible.validate(); err != nil {
		return err
	}

	return nil
}
