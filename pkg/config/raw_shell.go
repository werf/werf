package config

type rawShell struct {
	BeforeInstall             interface{} `yaml:"beforeInstall,omitempty"`
	Install                   interface{} `yaml:"install,omitempty"`
	BeforeSetup               interface{} `yaml:"beforeSetup,omitempty"`
	Setup                     interface{} `yaml:"setup,omitempty"`
	BuildArtifact             interface{} `yaml:"buildArtifact,omitempty"`
	CacheVersion              string      `yaml:"cacheVersion,omitempty"`
	BeforeInstallCacheVersion string      `yaml:"beforeInstallCacheVersion,omitempty"`
	InstallCacheVersion       string      `yaml:"installCacheVersion,omitempty"`
	BeforeSetupCacheVersion   string      `yaml:"beforeSetupCacheVersion,omitempty"`
	SetupCacheVersion         string      `yaml:"setupCacheVersion,omitempty"`
	BuildArtifactCacheVersion string      `yaml:"buildArtifactCacheVersion,omitempty"`

	rawDimg *rawDimg `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *rawShell) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := parentStack.Peek().(*rawDimg); ok {
		c.rawDimg = parent
	}

	type plain rawShell
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, c, c.rawDimg.doc); err != nil {
		return err
	}

	return nil
}

func (c *rawShell) toBaseDirective() (shellDimg *ShellDimg, err error) {
	shellDimg = &ShellDimg{}
	shellDimg.ShellBase = &ShellBase{}

	shellDimg.CacheVersion = c.CacheVersion
	shellDimg.BeforeInstallCacheVersion = c.BeforeInstallCacheVersion
	shellDimg.InstallCacheVersion = c.InstallCacheVersion
	shellDimg.BeforeSetupCacheVersion = c.BeforeSetupCacheVersion
	shellDimg.SetupCacheVersion = c.SetupCacheVersion

	if beforeInstall, err := InterfaceToStringArray(c.BeforeInstall, c, c.rawDimg.doc); err != nil {
		return nil, err
	} else {
		shellDimg.ShellBase.BeforeInstall = beforeInstall
	}

	if install, err := InterfaceToStringArray(c.Install, c, c.rawDimg.doc); err != nil {
		return nil, err
	} else {
		shellDimg.ShellBase.Install = install
	}

	if beforeSetup, err := InterfaceToStringArray(c.BeforeSetup, c, c.rawDimg.doc); err != nil {
		return nil, err
	} else {
		shellDimg.ShellBase.BeforeSetup = beforeSetup
	}

	if setup, err := InterfaceToStringArray(c.Setup, c, c.rawDimg.doc); err != nil {
		return nil, err
	} else {
		shellDimg.ShellBase.Setup = setup
	}

	shellDimg.ShellBase.raw = c

	return shellDimg, nil
}

func (c *rawShell) toDirective() (shellDimg *ShellDimg, err error) {
	shellDimg, err = c.toBaseDirective()
	if err != nil {
		return nil, err
	}

	if err := c.validateDirective(shellDimg); err != nil {
		return nil, err
	}

	return shellDimg, nil
}

func (c *rawShell) validateDirective(shellDimg *ShellDimg) error {
	if c.BuildArtifact != nil {
		return newDetailedConfigError("`buildArtifact` stage is not available for dimg, only for artifact!", c, c.rawDimg.doc)
	}

	if c.BuildArtifactCacheVersion != "" {
		return newDetailedConfigError("`buildArtifactCacheVersion` directive is not available for dimg, only for artifact!", c, c.rawDimg.doc)
	}

	if err := shellDimg.validate(); err != nil {
		return err
	}

	return nil
}

func (c *rawShell) toArtifactDirective() (shellArtifact *ShellArtifact, err error) {
	shellArtifact = &ShellArtifact{}

	if shellDimg, err := c.toBaseDirective(); err != nil {
		return nil, err
	} else {
		shellArtifact.ShellDimg = shellDimg
	}

	if buildArtifact, err := InterfaceToStringArray(c.BuildArtifact, c, c.rawDimg.doc); err != nil {
		return nil, err
	} else {
		shellArtifact.BuildArtifact = buildArtifact
	}

	shellArtifact.BuildArtifactCacheVersion = c.BuildArtifactCacheVersion

	if err := c.validateArtifactDirective(shellArtifact); err != nil {
		return nil, err
	}

	return shellArtifact, nil
}

func (c *rawShell) validateArtifactDirective(shellArtifact *ShellArtifact) error {
	if err := shellArtifact.validate(); err != nil {
		return err
	}

	return nil
}
