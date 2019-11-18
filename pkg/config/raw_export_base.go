package config

import (
	"path"
)

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
	exportBase.Add = path.Clean(c.Add)
	exportBase.To = path.Clean(c.To)

	if includePaths, err := InterfaceToStringArray(c.IncludePaths, c.rawOrigin.configSection(), c.rawOrigin.doc()); err != nil {
		return nil, err
	} else {
		for _, p := range includePaths {
			exportBase.IncludePaths = append(exportBase.IncludePaths, path.Clean(p))
		}
	}

	if excludePaths, err := InterfaceToStringArray(c.ExcludePaths, c.rawOrigin.configSection(), c.rawOrigin.doc()); err != nil {
		return nil, err
	} else {
		for _, p := range excludePaths {
			exportBase.ExcludePaths = append(exportBase.ExcludePaths, path.Clean(p))
		}
	}

	exportBase.Owner = c.Owner
	exportBase.Group = c.Group

	exportBase.raw = c

	if err := exportBase.validate(); err != nil {
		return nil, err
	}

	return exportBase, nil
}
