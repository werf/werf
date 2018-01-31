package config

type RawExportBase struct {
	Add          string      `yaml:"add,omitempty"`
	To           string      `yaml:"to,omitempty"`
	IncludePaths interface{} `yaml:"includePaths,omitempty"`
	ExcludePaths interface{} `yaml:"excludePaths,omitempty"`
	Owner        string      `yaml:"owner,omitempty"`
	Group        string      `yaml:"group,omitempty"`

	RawOrigin RawOrigin `yaml:"-"` // parent
}

func (c *RawExportBase) InlinedIntoRaw(RawOrigin RawOrigin) {
	c.RawOrigin = RawOrigin
}

func NewRawExportBase() RawExportBase {
	rawExportBase := RawExportBase{}
	rawExportBase.Add = "/"
	return rawExportBase
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

	exportBase.Raw = c

	if err := exportBase.Validate(); err != nil {
		return nil, err
	}

	return exportBase, nil
}
