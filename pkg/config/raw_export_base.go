package config

type RawExportBase struct {
	Add          string      `yaml:"add,omitempty"`
	To           string      `yaml:"to,omitempty"`
	IncludePaths interface{} `yaml:"includePaths,omitempty"`
	ExcludePaths interface{} `yaml:"excludePaths,omitempty"`
	Owner        string      `yaml:"owner,omitempty"`
	Group        string      `yaml:"group,omitempty"`
}

func (c *RawExportBase) ToDirective() (exportBase *ExportBase, err error) {
	exportBase = &ExportBase{}
	exportBase.Add = c.Add
	exportBase.To = c.To

	if includePaths, err := InterfaceToStringArray(c.IncludePaths); err != nil {
		return nil, err
	} else {
		exportBase.IncludePaths = includePaths
	}

	if excludePaths, err := InterfaceToStringArray(c.ExcludePaths); err != nil {
		return nil, err
	} else {
		exportBase.ExcludePaths = excludePaths
	}

	exportBase.Owner = c.Owner
	exportBase.Group = c.Group
	return exportBase, nil
}
