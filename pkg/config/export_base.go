package config

import (
	"fmt"

	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type ExportBase struct {
	Add          string
	To           string
	IncludePaths []string
	ExcludePaths []string
	Owner        string
	Group        string

	Raw *RawExportBase
}

func (c *ExportBase) Validate() error {
	if c.Add == "" || !IsAbsolutePath(c.Add) {
		return fmt.Errorf("`Add` required absolute path") // FIXME
	} else if c.To == "" || !IsAbsolutePath(c.To) {
		return fmt.Errorf("`To` required absolute path") // FIXME
	} else if !AllRelativePaths(c.IncludePaths) {
		return fmt.Errorf("`IncludePaths` should be relative paths") // FIXME
	} else if !AllRelativePaths(c.ExcludePaths) {
		return fmt.Errorf("`ExcludePaths` should be relative paths") // FIXME
	}
	return nil
}

func (c *ExportBase) ToRuby() ruby_marshal_config.ArtifactBaseExport {
	artifactBaseExport := ruby_marshal_config.ArtifactBaseExport{}
	artifactBaseExport.Cwd = c.Add
	artifactBaseExport.To = c.To
	artifactBaseExport.IncludePaths = c.IncludePaths
	artifactBaseExport.ExcludePaths = c.ExcludePaths
	artifactBaseExport.Owner = c.Owner
	artifactBaseExport.Group = c.Group
	return artifactBaseExport
}
