package config

import (
	"strings"
)

type GitExportBase struct {
	*GitExport
	StageDependencies *StageDependencies
}

func (c *ExportBase) GitMappingAdd() string {
	if c.Add == "/" {
		return ""
	}
	return strings.TrimPrefix(c.Add, "/")
}

func (c *ExportBase) GitMappingTo() string {
	return c.To
}
