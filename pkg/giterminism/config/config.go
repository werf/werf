package config

import (
	"github.com/werf/werf/pkg/giterminism"
	"github.com/werf/werf/pkg/giterminism_inspector/config"
)

type Config struct {
	config.GiterminismConfig
}

func NewConfig(projectDir string) (giterminism.Config, error) {
	c, err := config.PrepareConfig(projectDir)
	if err != nil {
		return nil, err
	}

	return Config{c}, err
}

func (c Config) IsUncommittedConfigAccepted() bool {
	return c.Config.AllowUncommitted
}

func (c Config) IsUncommittedConfigTemplateFileAccepted(path string) (bool, error) {
	return c.Config.IsUncommittedTemplateFileAccepted(path)
}

func (c Config) IsUncommittedConfigGoTemplateRenderingFileAccepted(path string) (bool, error) {
	return c.Config.GoTemplateRendering.IsUncommittedFileAccepted(path)
}
