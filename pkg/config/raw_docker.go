package config

type RawDocker struct {
	Volume     interface{}       `yaml:"VOLUME,omitempty"`
	Expose     interface{}       `yaml:"EXPOSE,omitempty"`
	Env        map[string]string `yaml:"ENV,omitempty"`
	Label      map[string]string `yaml:"LABEL,omitempty"`
	Cmd        interface{}       `yaml:"CMD,omitempty"`
	Onbuild    interface{}       `yaml:"ONBUILD,omitempty"`
	Workdir    string            `yaml:"WORKDIR,omitempty"`
	User       string            `yaml:"USER,omitempty"`
	Entrypoint interface{}       `yaml:"ENTRYPOINT,omitempty"`

	RawDimg *RawDimg `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *RawDocker) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := ParentStack.Peek().(*RawDimg); ok {
		c.RawDimg = parent
	}

	type plain RawDocker
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := CheckOverflow(c.UnsupportedAttributes, c, c.RawDimg.Doc); err != nil {
		return err
	}

	return nil
}

func (c *RawDocker) ToDirective() (docker *Docker, err error) {
	docker = &Docker{}

	if volume, err := InterfaceToStringArray(c.Volume, c, c.RawDimg.Doc); err != nil {
		return nil, err
	} else {
		docker.Volume = volume
	}

	if expose, err := InterfaceToStringArray(c.Expose, c, c.RawDimg.Doc); err != nil {
		return nil, err
	} else {
		docker.Expose = expose
	}

	docker.Env = c.Env
	docker.Label = c.Label

	if cmd, err := InterfaceToStringArray(c.Cmd, c, c.RawDimg.Doc); err != nil {
		return nil, err
	} else {
		docker.Cmd = cmd
	}

	if onbuild, err := InterfaceToStringArray(c.Onbuild, c, c.RawDimg.Doc); err != nil {
		return nil, err
	} else {
		docker.Onbuild = onbuild
	}

	docker.Workdir = c.Workdir
	docker.User = c.User

	if entrypoint, err := InterfaceToStringArray(c.Entrypoint, c, c.RawDimg.Doc); err != nil {
		return nil, err
	} else {
		docker.Entrypoint = entrypoint
	}

	docker.Raw = c

	if err := c.ValidateDirective(docker); err != nil {
		return nil, err
	}

	return docker, nil
}

func (c *RawDocker) ValidateDirective(docker *Docker) (err error) {
	if err := docker.Validate(); err != nil {
		return err
	}

	return nil
}
