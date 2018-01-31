package config

type RawCookbook struct {
	Name              string                 `yaml:"name"`
	VersionConstraint string                 `yaml:"versionConstraint"`
	Path              string                 `yaml:"path"`
	Fields            map[string]interface{} `yaml:",inline"`

	RawChef *RawChef `yaml:"-"` // parent
}

func (c *RawCookbook) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := ParentStack.Peek().(*RawChef); ok {
		c.RawChef = parent
	}

	ParentStack.Push(c)
	type plain RawCookbook
	err := unmarshal((*plain)(c))
	ParentStack.Pop()
	if err != nil {
		return err
	}

	return nil
}
