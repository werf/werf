package config

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/werf/logboek"
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
	exportBase.Add = filepath.ToSlash(filepath.Clean(c.Add))

	if strings.HasSuffix(c.To, "/") && c.To != "/" {
		toWithoutTrailingSlash := strings.TrimSuffix(c.To, "/")
		logboek.Context(context.Background()).Warn().LogF(
			"WARNING: `to: %s` will be treated like `to: %s`, i.e. file/directory from `add: %s` will NOT be copied inside of the %q, instead it will be copied as %q! To hide this warning, change `to: %s` to `to: %s`.\n",
			c.To, toWithoutTrailingSlash, c.Add, c.To, toWithoutTrailingSlash, c.To, toWithoutTrailingSlash,
		)
	}
	exportBase.To = filepath.ToSlash(filepath.Clean(c.To))

	if includePaths, err := InterfaceToStringArray(c.IncludePaths, c.rawOrigin.configSection(), c.rawOrigin.doc()); err != nil {
		return nil, err
	} else {
		for _, p := range includePaths {
			exportBase.IncludePaths = append(exportBase.IncludePaths, filepath.ToSlash(filepath.Clean(p)))
		}
	}

	if excludePaths, err := InterfaceToStringArray(c.ExcludePaths, c.rawOrigin.configSection(), c.rawOrigin.doc()); err != nil {
		return nil, err
	} else {
		for _, p := range excludePaths {
			exportBase.ExcludePaths = append(exportBase.ExcludePaths, filepath.ToSlash(filepath.Clean(p)))
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
