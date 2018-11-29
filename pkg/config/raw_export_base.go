package config

type rawExportBase struct {
	Add          string      `yaml:"add,omitempty"`
	To           string      `yaml:"to,omitempty"`
	IncludePaths interface{} `yaml:"includePaths,omitempty"`
	ExcludePaths interface{} `yaml:"excludePaths,omitempty"`
	Owner        string      `yaml:"owner,omitempty"`
	Group        string      `yaml:"group,omitempty"`

	rawOrigin rawOrigin `yaml:"-"` // parent
}

func (c *rawExportBase) inlinedIntoRaw(rawOrigin rawOrigin) {
	c.rawOrigin = rawOrigin
}

func (c *rawExportBase) toDirective() (exportBase *ExportBase, err error) {
	exportBase = &ExportBase{}
	exportBase.Add = c.Add
	exportBase.To = c.To

	if includePaths, err := InterfaceToStringArray(c.IncludePaths, c.rawOrigin.configSection(), c.rawOrigin.doc()); err != nil {
		return nil, err
	} else {
		exportBase.IncludePaths = includePaths
	}

	if excludePaths, err := InterfaceToStringArray(c.ExcludePaths, c.rawOrigin.configSection(), c.rawOrigin.doc()); err != nil {
		return nil, err
	} else {
		exportBase.ExcludePaths = excludePaths
	}

	exportBase.Owner = c.Owner
	exportBase.Group = c.Group

	exportBase.raw = c

	if err := exportBase.validate(); err != nil {
		return nil, err
	}

	return exportBase, nil
}
