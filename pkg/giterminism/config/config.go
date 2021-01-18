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

func (c Config) IsConfigGoTemplateRenderingEnvNameAccepted(envName string) (bool, error) {
	return c.Config.GoTemplateRendering.IsEnvNameAccepted(envName)
}

func (c Config) IsConfigStapelFromLatestAccepted() bool {
	return c.Config.Stapel.AllowFromLatest
}

func (c Config) IsConfigStapelGitBranchAccepted() bool {
	return c.Config.Stapel.Git.AllowBranch
}

func (c Config) IsConfigStapelMountBuildDirAccepted() bool {
	return c.Config.Stapel.Mount.AllowBuildDir
}

func (c Config) IsConfigStapelMountFromPathAccepted(fromPath string) (bool, error) {
	return c.Config.Stapel.Mount.IsFromPathAccepted(fromPath)
}

func (c Config) IsConfigDockerfileContextAddFileAccepted(relPath string) (bool, error) {
	return c.Config.Dockerfile.IsContextAddFileAccepted(relPath)
}

func (c Config) IsUncommittedDockerfileAccepted(path string) (bool, error) {
	return c.Config.Dockerfile.IsUncommittedAccepted(path)
}

func (c Config) IsUncommittedDockerignoreAccepted(path string) (bool, error) {
	return c.Config.Dockerfile.IsUncommittedDockerignoreAccepted(path)
}
