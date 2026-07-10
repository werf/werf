package config

import (
	"context"

	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

type rawImport struct {
	From   string `yaml:"from,omitempty"`
	Image  string `yaml:"image,omitempty"` // Deprecated: use `from` instead.
	Before string `yaml:"before,omitempty"`
	After  string `yaml:"after,omitempty"`

	rawExport      `yaml:",inline"`
	rawStapelImage *rawStapelImage `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *rawImport) configSection() interface{} {
	return c
}

func (c *rawImport) doc() *doc {
	return c.rawStapelImage.doc
}

func (c *rawImport) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := parentStack.Peek().(*rawStapelImage); ok {
		c.rawStapelImage = parent
	}

	parentStack.Push(c)
	type plain rawImport
	err := unmarshal((*plain)(c))
	parentStack.Pop()
	if err != nil {
		return err
	}

	if c.Image != "" {
		if c.From != "" {
			return newDetailedConfigError("specify only `from: NAME` or deprecated `image: NAME` for import, not both!", c, c.rawStapelImage.doc)
		}
		global_warnings.GlobalDeprecationWarningLn(context.Background(), "`image: NAME` for import is deprecated and will be removed in a future version, use `from: NAME` instead.")
		c.From = c.Image
	}

	c.rawExport.inlinedIntoRaw(c)

	if err := checkOverflow(c.UnsupportedAttributes, c, c.rawStapelImage.doc); err != nil {
		return err
	}

	if c.rawExport.rawExportBase.To == "" {
		c.rawExport.rawExportBase.To = c.rawExport.rawExportBase.Add
	}

	return nil
}

func (c *rawImport) toDirective() (imp *Import, err error) {
	imp = &Import{}

	if export, tempErr := c.rawExport.toDirective(); tempErr != nil {
		return nil, tempErr
	} else {
		imp.Export = export
	}

	if c.From != "" {
		imp.From = c.From
	}

	imp.Before = c.Before
	imp.After = c.After

	imp.raw = c

	if err = c.validateDirective(imp); err != nil {
		return nil, err
	}

	return imp, nil
}

func (c *rawImport) validateDirective(imp *Import) (err error) {
	if err = imp.validate(); err != nil {
		return err
	}

	return nil
}
