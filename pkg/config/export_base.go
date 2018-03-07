package config

import "github.com/flant/dapp/pkg/config/ruby_marshal_config"

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
		return NewDetailedConfigError("`add: PATH` absolute path required for import!", c.Raw.RawOrigin.ConfigSection(), c.Raw.RawOrigin.Doc())
	} else if c.To == "" || !IsAbsolutePath(c.To) {
		return NewDetailedConfigError("`to: PATH` absolute path required for import!", c.Raw.RawOrigin.ConfigSection(), c.Raw.RawOrigin.Doc())
	} else if !AllRelativePaths(c.IncludePaths) {
		return NewDetailedConfigError("`includePaths: [PATH, ...]|PATH` should be relative paths!", c.Raw.RawOrigin.ConfigSection(), c.Raw.RawOrigin.Doc())
	} else if !AllRelativePaths(c.ExcludePaths) {
		return NewDetailedConfigError("`excludePaths: [PATH, ...]|PATH` should be relative paths!", c.Raw.RawOrigin.ConfigSection(), c.Raw.RawOrigin.Doc())
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
