package config

type RawAnsible struct {
	BeforeInstall             []RawAnsibleTask `yaml:"beforeInstall"`
	Install                   []RawAnsibleTask `yaml:"install"`
	BeforeSetup               []RawAnsibleTask `yaml:"beforeSetup"`
	Setup                     []RawAnsibleTask `yaml:"setup"`
	BuildArtifact             []RawAnsibleTask `yaml:"buildArtifact"`
	CacheVersion              string           `yaml:"cacheVersion"`
	BeforeInstallCacheVersion string           `yaml:"beforeInstallCacheVersion"`
	InstallCacheVersion       string           `yaml:"installCacheVersion"`
	BeforeSetupCacheVersion   string           `yaml:"beforeSetupCacheVersion"`
	SetupCacheVersion         string           `yaml:"setupCacheVersion"`
	BuildArtifactCacheVersion string           `yaml:"buildArtifactCacheVersion"`

	RawDimg *RawDimg `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *RawAnsible) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := ParentStack.Peek().(*RawDimg); ok {
		c.RawDimg = parent
	}

	ParentStack.Push(c)
	type plain RawAnsible
	err := unmarshal((*plain)(c))
	ParentStack.Pop()
	if err != nil {
		return err
	}

	if err := CheckOverflow(c.UnsupportedAttributes, c, c.RawDimg.Doc); err != nil {
		return err
	}

	return nil
}

func (c *RawAnsible) ToDirective() (ansible *Ansible, err error) {
	ansible = &Ansible{}

	ansible.CacheVersion = c.CacheVersion
	ansible.BeforeInstallCacheVersion = c.BeforeInstallCacheVersion
	ansible.InstallCacheVersion = c.InstallCacheVersion
	ansible.BeforeSetupCacheVersion = c.BeforeSetupCacheVersion
	ansible.SetupCacheVersion = c.SetupCacheVersion
	ansible.BuildArtifactCacheVersion = c.BuildArtifactCacheVersion

	for ind := range c.BeforeInstall {
		if ansibleTask, err := c.BeforeInstall[ind].ToDirective(); err != nil {
			return nil, err
		} else {
			ansible.BeforeInstall = append(ansible.BeforeInstall, ansibleTask)
		}
	}

	for ind := range c.Install {
		if ansibleTask, err := c.Install[ind].ToDirective(); err != nil {
			return nil, err
		} else {
			ansible.Install = append(ansible.Install, ansibleTask)
		}
	}

	for ind := range c.BeforeSetup {
		if ansibleTask, err := c.BeforeSetup[ind].ToDirective(); err != nil {
			return nil, err
		} else {
			ansible.BeforeSetup = append(ansible.BeforeSetup, ansibleTask)
		}
	}

	for ind := range c.Setup {
		if ansibleTask, err := c.Setup[ind].ToDirective(); err != nil {
			return nil, err
		} else {
			ansible.Setup = append(ansible.Setup, ansibleTask)
		}
	}

	for ind := range c.BuildArtifact {
		if ansibleTask, err := c.BuildArtifact[ind].ToDirective(); err != nil {
			return nil, err
		} else {
			ansible.BuildArtifact = append(ansible.BuildArtifact, ansibleTask)
		}
	}

	ansible.Raw = c

	if err := c.ValidateDirective(ansible); err != nil {
		return nil, err
	}

	return ansible, nil
}

func (c *RawAnsible) ValidateDirective(ansible *Ansible) (err error) {
	if err := ansible.Validate(); err != nil {
		return err
	}

	return nil
}
