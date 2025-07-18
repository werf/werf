package config

type rawMetaBuild struct {
	CacheVersion string              `yaml:"cacheVersion,omitempty"`
	Platform     []string            `yaml:"platform,omitempty"`
	Staged       bool                `yaml:"staged,omitempty"`
	RawImageSpec *rawImageSpecGlobal `yaml:"imageSpec,omitempty"`
	rawMeta      *rawMeta

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *rawMetaBuild) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := parentStack.Peek().(*rawMeta); ok {
		c.rawMeta = parent
	}

	parentStack.Push(c)
	type plain rawMetaBuild
	err := unmarshal((*plain)(c))
	parentStack.Pop()
	if err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, nil, c.rawMeta.doc); err != nil {
		return err
	}

	return nil
}

func (c *rawMetaBuild) toMetaBuild() MetaBuild {
	metaBuild := MetaBuild{}
	metaBuild.CacheVersion = c.CacheVersion
	metaBuild.Platform = c.Platform
	metaBuild.Staged = c.Staged
	if c.RawImageSpec != nil {
		metaBuild.ImageSpec = c.RawImageSpec.toDirective()
	}
	return metaBuild
}
