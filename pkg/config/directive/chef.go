package config

type Chef struct {
	Cookbook   string          `yaml:"cookbook,omitempty"`
	Recipe     interface{}     `yaml:"recipe,omitempty"`
	Attributes *ChefAttributes `yaml:"attributes,omitempty"`

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}
type ChefAttributes map[interface{}]interface{}

func (c *Chef) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Chef
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := CheckOverflow(c.UnsupportedAttributes, c); err != nil {
		return err
	}

	return nil
}
