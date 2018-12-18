package config

import "fmt"

type rawDocker struct {
	Volume      interface{}       `yaml:"VOLUME,omitempty"`
	Expose      interface{}       `yaml:"EXPOSE,omitempty"`
	Env         map[string]string `yaml:"ENV,omitempty"`
	Label       map[string]string `yaml:"LABEL,omitempty"`
	Cmd         interface{}       `yaml:"CMD,omitempty"`
	Onbuild     interface{}       `yaml:"ONBUILD,omitempty"`
	Workdir     string            `yaml:"WORKDIR,omitempty"`
	User        string            `yaml:"USER,omitempty"`
	Entrypoint  interface{}       `yaml:"ENTRYPOINT,omitempty"`
	StopSignal  interface{}       `yaml:"STOPSIGNAL,omitempty"`
	HealthCheck string            `yaml:"HEALTHCHECK,omitempty"`

	rawDimg *rawDimg `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *rawDocker) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := parentStack.Peek().(*rawDimg); ok {
		c.rawDimg = parent
	}

	type plain rawDocker
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, c, c.rawDimg.doc); err != nil {
		return err
	}

	return nil
}

func (c *rawDocker) toDirective() (docker *Docker, err error) {
	docker = &Docker{}

	if volume, err := InterfaceToStringArray(c.Volume, c, c.rawDimg.doc); err != nil {
		return nil, err
	} else {
		docker.Volume = volume
	}

	if expose, err := InterfaceToStringArray(c.Expose, c, c.rawDimg.doc); err != nil {
		return nil, err
	} else {
		docker.Expose = expose
	}

	docker.Env = c.Env
	docker.Label = c.Label

	if cmd, err := InterfaceToStringArray(c.Cmd, c, c.rawDimg.doc); err != nil {
		return nil, err
	} else {
		docker.Cmd = cmd
	}

	if onbuild, err := InterfaceToStringArray(c.Onbuild, c, c.rawDimg.doc); err != nil {
		return nil, err
	} else {
		docker.Onbuild = onbuild
	}

	docker.Workdir = c.Workdir
	docker.User = c.User

	if entrypoint, err := InterfaceToStringArray(c.Entrypoint, c, c.rawDimg.doc); err != nil {
		return nil, err
	} else {
		docker.Entrypoint = entrypoint
	}

	if c.StopSignal != nil {
		docker.StopSignal = fmt.Sprintf("%v", c.StopSignal)
	}

	docker.HealthCheck = c.HealthCheck

	docker.raw = c

	if err := c.validateDirective(docker); err != nil {
		return nil, err
	}

	return docker, nil
}

func (c *rawDocker) validateDirective(docker *Docker) (err error) {
	if err := docker.validate(); err != nil {
		return err
	}

	return nil
}
