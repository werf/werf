package config

type Docker struct {
	Volume     interface{}       `yaml:"VOLUME,omitempty"`
	Expose     interface{}       `yaml:"EXPOSE,omitempty"`
	Env        map[string]string `yaml:"ENV,omitempty"`
	Label      map[string]string `yaml:"LABEL,omitempty"`
	Cmd        interface{}       `yaml:"CMD,omitempty"`
	Onbuild    interface{}       `yaml:"ONBUILD,omitempty"`
	Workdir    string            `yaml:"WORKDIR,omitempty"`
	User       string            `yaml:"USER,omitempty"`
	Entrypoint interface{}       `yaml:"ENTRYPOINT,omitempty"`

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *Docker) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Docker
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := CheckOverflow(c.UnsupportedAttributes, c); err != nil {
		return err
	}

	return nil
}
