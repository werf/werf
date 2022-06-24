package config

import (
	"fmt"
	"strings"
)

type rawDocker struct {
	Volume      interface{}       `yaml:"VOLUME,omitempty"`
	Expose      interface{}       `yaml:"EXPOSE,omitempty"`
	Env         map[string]string `yaml:"ENV,omitempty"`
	Label       map[string]string `yaml:"LABEL,omitempty"`
	Cmd         interface{}       `yaml:"CMD,omitempty"`
	Workdir     string            `yaml:"WORKDIR,omitempty"`
	User        string            `yaml:"USER,omitempty"`
	Entrypoint  interface{}       `yaml:"ENTRYPOINT,omitempty"`
	HealthCheck string            `yaml:"HEALTHCHECK,omitempty"`

	ExactValues bool `yaml:"exactValues,omitempty"`

	rawStapelImage *rawStapelImage `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *rawDocker) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := parentStack.Peek().(*rawStapelImage); ok {
		c.rawStapelImage = parent
	}

	type plain rawDocker
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, c, c.rawStapelImage.doc); err != nil {
		return err
	}

	return nil
}

func (c *rawDocker) toDirective() (docker *Docker, err error) {
	docker = &Docker{}

	if volume, err := InterfaceToStringArray(c.Volume, c, c.rawStapelImage.doc); err != nil {
		return nil, err
	} else {
		docker.Volume = volume
	}

	if expose, err := InterfaceToStringArray(c.Expose, c, c.rawStapelImage.doc); err != nil {
		return nil, err
	} else {
		docker.Expose = expose
	}

	docker.Env = c.Env
	docker.Label = c.Label

	if cmd, err := prepareCommand(c.Cmd, c, c.rawStapelImage.doc); err != nil {
		return nil, err
	} else {
		docker.Cmd = cmd
	}

	docker.Workdir = c.Workdir
	docker.User = c.User

	if entrypoint, err := prepareCommand(c.Entrypoint, c, c.rawStapelImage.doc); err != nil {
		return nil, err
	} else {
		docker.Entrypoint = entrypoint
	}

	if docker.Entrypoint != "" && docker.Cmd == "" {
		docker.Cmd = "[]"
	}

	docker.HealthCheck = c.HealthCheck
	docker.ExactValues = c.ExactValues

	docker.raw = c

	if err := c.validateDirective(docker); err != nil {
		return nil, err
	}

	return docker, nil
}

func prepareCommand(stringOrArray, configSection interface{}, doc *doc) (cmd string, err error) {
	if stringOrArray != nil {
		if val, ok := stringOrArray.(string); ok {
			cmd = val
		} else if interfaceArray, ok := stringOrArray.([]interface{}); ok {
			var stringArray []string
			for _, interf := range interfaceArray {
				if val, ok := interf.(string); ok {
					stringArray = append(stringArray, val)
				} else {
					return cmd, newDetailedConfigError(fmt.Sprintf("single string or array of strings expected, got `%v`!", stringOrArray), configSection, doc)
				}
			}

			if len(stringArray) == 0 {
				cmd = "[]"
			} else {
				cmd = fmt.Sprintf("[\"%s\"]", strings.Join(stringArray, "\", \""))
			}
		} else {
			return cmd, newDetailedConfigError(fmt.Sprintf("single string or array of strings expected, got `%v`!", stringOrArray), configSection, doc)
		}
	}

	return
}

func (c *rawDocker) validateDirective(docker *Docker) (err error) {
	if err := docker.validate(); err != nil {
		return err
	}

	return nil
}
