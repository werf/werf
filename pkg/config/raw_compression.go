package config

import (
	"github.com/werf/werf/v3/pkg/image"
)

type rawCompression struct {
	Format string `yaml:"format,omitempty"`
	Level  *int   `yaml:"level,omitempty"`

	doc *doc `yaml:"-"`

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *rawCompression) UnmarshalYAML(unmarshal func(interface{}) error) error {
	switch parent := parentStack.Peek().(type) {
	case *rawMetaBuild:
		c.doc = parent.rawMeta.doc
	case *rawStapelImage:
		c.doc = parent.doc
	case *rawImageFromDockerfile:
		c.doc = parent.doc
	}

	parentStack.Push(c)
	type plain rawCompression
	err := unmarshal((*plain)(c))
	parentStack.Pop()
	if err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, nil, c.doc); err != nil {
		return err
	}

	if err := c.toDirective().Validate(); err != nil {
		return newDetailedConfigError("invalid `compression`: "+err.Error(), c, c.doc)
	}

	return nil
}

func (c *rawCompression) toDirective() image.LayerCompression {
	if c == nil {
		return image.LayerCompression{}
	}

	return image.LayerCompression{Format: c.Format, Level: c.Level}
}
