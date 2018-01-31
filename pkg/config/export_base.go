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
		return fmt.Errorf("`add: PATH` absolute path required for import!\n\n%s\n%s", DumpConfigSection(c.Raw.RawOrigin.ConfigSection()), DumpConfigDoc(c.Raw.RawOrigin.Doc()))
	} else if c.To == "" || !IsAbsolutePath(c.To) {
		return fmt.Errorf("`to: PATH` absolute path required for import!\n\n%s\n%s", DumpConfigSection(c.Raw.RawOrigin.ConfigSection()), DumpConfigDoc(c.Raw.RawOrigin.Doc()))
	} else if !AllRelativePaths(c.IncludePaths) {
		return fmt.Errorf("`includePaths: [PATH, ...]|PATH` should be relative paths\n\n%s\n%s", DumpConfigSection(c.Raw.RawOrigin.ConfigSection()), DumpConfigDoc(c.Raw.RawOrigin.Doc()))
	} else if !AllRelativePaths(c.ExcludePaths) {
		return fmt.Errorf("`excludePaths: [PATH, ...]|PATH` should be relative paths\n\n%s\n%s", DumpConfigSection(c.Raw.RawOrigin.ConfigSection()), DumpConfigDoc(c.Raw.RawOrigin.Doc()))
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
