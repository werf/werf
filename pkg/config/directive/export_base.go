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
}

func (c *ExportBase) Validate() error {
	if c.To == "" {
		return fmt.Errorf("to не может быть пустым!") // FIXME
	}

	// TODO: валидация `Add`, `To` абсолютные пути
	// TODO: валидация `IncludePaths`, `ExcludePaths` относительные

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
