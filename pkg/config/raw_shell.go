package config

type RawShell struct {
	BeforeInstall             interface{} `yaml:"beforeInstall,omitempty"`
	Install                   interface{} `yaml:"install,omitempty"`
	BeforeSetup               interface{} `yaml:"beforeSetup,omitempty"`
	Setup                     interface{} `yaml:"setup,omitempty"`
	BuildArtifact             interface{} `yaml:"buildArtifact,omitempty"`
	CacheVersion              string      `yaml:"cacheVersion"`
	BeforeInstallCacheVersion string      `yaml:"beforeInstallCacheVersion"`
	InstallCacheVersion       string      `yaml:"installCacheVersion"`
	BeforeSetupCacheVersion   string      `yaml:"beforeSetupCacheVersion"`
	SetupCacheVersion         string      `yaml:"setupCacheVersion"`
	BuildArtifactCacheVersion string      `yaml:"buildArtifactCacheVersion"`

	RawDimg *RawDimg `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *RawShell) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := ParentStack.Peek().(*RawDimg); ok {
		c.RawDimg = parent
	}

	type plain RawShell
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := CheckOverflow(c.UnsupportedAttributes, c, c.RawDimg.Doc); err != nil {
		return err
	}

	return nil
}

func (c *RawShell) ToBaseDirective() (shellDimg *ShellDimg, err error) {
	shellDimg = &ShellDimg{}
	shellDimg.ShellBase = &ShellBase{}

	shellDimg.CacheVersion = c.CacheVersion
	shellDimg.BeforeInstallCacheVersion = c.BeforeInstallCacheVersion
	shellDimg.InstallCacheVersion = c.InstallCacheVersion
	shellDimg.BeforeSetupCacheVersion = c.BeforeSetupCacheVersion
	shellDimg.SetupCacheVersion = c.SetupCacheVersion

	if beforeInstall, err := InterfaceToStringArray(c.BeforeInstall, c, c.RawDimg.Doc); err != nil {
		return nil, err
	} else {
		shellDimg.ShellBase.BeforeInstall = beforeInstall
	}

	if install, err := InterfaceToStringArray(c.Install, c, c.RawDimg.Doc); err != nil {
		return nil, err
	} else {
		shellDimg.ShellBase.Install = install
	}

	if beforeSetup, err := InterfaceToStringArray(c.BeforeSetup, c, c.RawDimg.Doc); err != nil {
		return nil, err
	} else {
		shellDimg.ShellBase.BeforeSetup = beforeSetup
	}

	if setup, err := InterfaceToStringArray(c.Setup, c, c.RawDimg.Doc); err != nil {
		return nil, err
	} else {
		shellDimg.ShellBase.Setup = setup
	}

	shellDimg.ShellBase.Raw = c

	return shellDimg, nil
}

func (c *RawShell) ToDirective() (shellDimg *ShellDimg, err error) {
	shellDimg, err = c.ToBaseDirective()
	if err != nil {
		return nil, err
	}

	if err := c.ValidateDirective(shellDimg); err != nil {
		return nil, err
	}

	return shellDimg, nil
}

func (c *RawShell) ValidateDirective(shellDimg *ShellDimg) error {
	if c.BuildArtifact != nil {
		return NewDetailedConfigError("`buildArtifact` stage is not available for dimg, only for artifact!", c, c.RawDimg.Doc)
	}

	if c.BuildArtifactCacheVersion != "" {
		return NewDetailedConfigError("`buildArtifactCacheVersion` directive is not available for dimg, only for artifact!", c, c.RawDimg.Doc)
	}

	if err := shellDimg.Validate(); err != nil {
		return err
	}

	return nil
}

func (c *RawShell) ToArtifactDirective() (shellArtifact *ShellArtifact, err error) {
	shellArtifact = &ShellArtifact{}

	if shellDimg, err := c.ToBaseDirective(); err != nil {
		return nil, err
	} else {
		shellArtifact.ShellDimg = shellDimg
	}

	shellArtifact.BuildArtifactCacheVersion = c.BuildArtifactCacheVersion

	if err := c.ValidateArtifactDirective(shellArtifact); err != nil {
		return nil, err
	}

	return shellArtifact, nil
}

func (c *RawShell) ValidateArtifactDirective(shellArtifact *ShellArtifact) error {
	if err := shellArtifact.Validate(); err != nil {
		return err
	}

	return nil
}
