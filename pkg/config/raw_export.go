package config

type rawExport struct {
	rawExportBase `yaml:",inline"`

	rawOrigin rawOrigin `yaml:"-"` // parent
}

func (c *rawExport) inlinedIntoRaw(rawOrigin rawOrigin) {
	c.rawOrigin = rawOrigin
	c.rawExportBase.inlinedIntoRaw(rawOrigin)
}

func (c *rawExport) toDirective() (export *Export, err error) {
	export = &Export{}

	if exportBase, err := c.rawExportBase.toDirective(); err != nil {
		return nil, err
	} else {
		export.ExportBase = exportBase
	}

	export.raw = c

	if err := export.validate(); err != nil {
		return nil, err
	}

	return export, nil
}
